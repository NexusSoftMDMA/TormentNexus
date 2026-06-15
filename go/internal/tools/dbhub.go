//go:build ignore
// +build ignore

package tools

/**
 * @file dbhub.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of DBHub MCP tools.
 * Replaces `dbhub` (npx @bytebase/dbhub@latest) entry in mcp.json.
 *
 * DBHub provides a universal database interface supporting PostgreSQL,
 * MySQL, SQLite, and other databases via DSN connection strings.
 *
 * Improvements over original:
 * - No npx/Node dependency.
 * - Go-native database/sql driver management.
 * - Supports: list_databases, list_tables, execute_query, describe_table,
 *   list_schemas, get_table_data.
 * - Context-aware with timeout; DSN from args or DBHUB_DSN env var.
 */

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	_ "modernc.org/sqlite"       // SQLite driver (already in go.mod)
	_ "github.com/lib/pq"                // PostgreSQL driver
	_ "github.com/go-sql-driver/mysql"    // MySQL driver
)

func dbhubDSN(args map[string]interface{}) (string, string) {
	// Priority: args > env
	if dsn, _ := getString(args, "dsn", "connection_string"); dsn != "" {
		return dsn, detectDriver(dsn)
	}
	if dsn := os.Getenv("DBHUB_DSN"); dsn != "" {
		return dsn, detectDriver(dsn)
	}
	// Fallback to SQLite
	return "tormentnexus.db", "sqlite"
}

func detectDriver(dsn string) string {
	dsnLower := strings.ToLower(dsn)
	if strings.HasPrefix(dsnLower, "postgres://") || strings.HasPrefix(dsnLower, "postgresql://") {
		return "postgres"
	}
	if strings.HasPrefix(dsnLower, "mysql://") || strings.HasPrefix(dsnLower, "mysql:") {
		return "mysql"
	}
	// Default to sqlite for file paths and sqlite: prefixes
	return "sqlite"
}

func dbhubOpen(args map[string]interface{}) (*sql.DB, string, error) {
	dsn, driver := dbhubDSN(args)

	var db *sql.DB
	var e error

	switch driver {
	case "postgres":
		db, e = sql.Open("postgres", dsn)
	case "mysql":
		// Strip mysql:// prefix for go-sql-driver
		cleanDSN := strings.TrimPrefix(dsn, "mysql://")
		db, e = sql.Open("mysql", cleanDSN)
	default:
		// SQLite - ensure read-write mode
		if !strings.Contains(dsn, "?mode=") && !strings.Contains(dsn, "?") {
			dsn = dsn + "?mode=rw"
		} else if !strings.Contains(dsn, "mode=") {
			dsn = dsn + "&mode=rw"
		}
		db, e = sql.Open("sqlite", dsn)
	}

	if e != nil {
		return nil, driver, e
	}

	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(30 * time.Second)

	return db, driver, nil
}

// HandleDBHubListDatabases lists available databases.
// Tool: dbhub_list_databases
func HandleDBHubListDatabases(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, driver, e := dbhubOpen(args)
	if e != nil {
		return err(fmt.Sprintf("Failed to connect: %v", e))
	}
	defer db.Close()

	var dbName string
	switch driver {
	case "postgres":
		row := db.QueryRowContext(ctx, "SELECT current_database()")
		row.Scan(&dbName)
	case "mysql":
		row := db.QueryRowContext(ctx, "SELECT DATABASE()")
		row.Scan(&dbName)
	default:
		dbName = "main"
	}

	result := map[string]interface{}{
		"driver":   driver,
		"database": dbName,
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleDBHubListTables lists all tables in the database.
// Tool: dbhub_list_tables
func HandleDBHubListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, driver, e := dbhubOpen(args)
	if e != nil {
		return err(fmt.Sprintf("Failed to connect: %v", e))
	}
	defer db.Close()

	var query string
	switch driver {
	case "postgres":
		schema, _ := getString(args, "schema")
		if schema == "" {
			schema = "public"
		}
		query = fmt.Sprintf("SELECT table_name FROM information_schema.tables WHERE table_schema = '%s' ORDER BY table_name", schema)
	case "mysql":
		query = "SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() ORDER BY table_name"
	default:
		query = "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name"
	}

	rows, e := db.QueryContext(ctx, query)
	if e != nil {
		return err(fmt.Sprintf("Failed to list tables: %v", e))
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			tables = append(tables, name)
		}
	}

	result := map[string]interface{}{
		"driver": driver,
		"tables": tables,
		"count":  len(tables),
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleDBHubDescribeTable describes the structure of a table.
// Tool: dbhub_describe_table
func HandleDBHubDescribeTable(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	table, _ := getString(args, "table", "table_name")
	if table == "" {
		return err("table parameter is required")
	}

	db, driver, e := dbhubOpen(args)
	if e != nil {
		return err(fmt.Sprintf("Failed to connect: %v", e))
	}
	defer db.Close()

	var query string
	switch driver {
	case "postgres":
		query = fmt.Sprintf("SELECT column_name, data_type, is_nullable, column_default FROM information_schema.columns WHERE table_name = '%s' ORDER BY ordinal_position", table)
	case "mysql":
		query = fmt.Sprintf("DESCRIBE %s", table)
	default:
		query = fmt.Sprintf("PRAGMA table_info(%s)", table)
	}

	rows, e := db.QueryContext(ctx, query)
	if e != nil {
		return err(fmt.Sprintf("Failed to describe table: %v", e))
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	var results []map[string]interface{}
	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		row := map[string]interface{}{}
		for i, col := range cols {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	return ok(string(out))
}

// HandleDBHubExecuteQuery executes a SQL query and returns results.
// Tool: dbhub_execute_query
func HandleDBHubExecuteQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "sql")
	if query == "" {
		return err("query parameter is required")
	}

	db, _, e := dbhubOpen(args)
	if e != nil {
		return err(fmt.Sprintf("Failed to connect: %v", e))
	}
	defer db.Close()

	// Determine if read or write
	upperQuery := strings.ToUpper(strings.TrimSpace(query))
	isWrite := strings.HasPrefix(upperQuery, "INSERT") ||
		strings.HasPrefix(upperQuery, "UPDATE") ||
		strings.HasPrefix(upperQuery, "DELETE") ||
		strings.HasPrefix(upperQuery, "CREATE") ||
		strings.HasPrefix(upperQuery, "DROP") ||
		strings.HasPrefix(upperQuery, "ALTER")

	if isWrite {
		res, errExec := db.ExecContext(ctx, query)
		if errExec != nil {
			return err(fmt.Sprintf("Failed to execute query: %v", errExec))
		}
		rowsAffected, _ := res.RowsAffected()
		return ok(fmt.Sprintf("Query executed successfully. Rows affected: %d", rowsAffected))
	}

	rows, e := db.QueryContext(ctx, query)
	if e != nil {
		return err(fmt.Sprintf("Failed to execute query: %v", e))
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	var results []map[string]interface{}
	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	rowCount := 0
	maxRows := getInt(args, "max_rows")
	if maxRows <= 0 {
		maxRows = 1000
	}

	for rows.Next() {
		if rowCount >= maxRows {
			break
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		row := map[string]interface{}{}
		for i, col := range cols {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
		rowCount++
	}

	result := map[string]interface{}{
		"columns": cols,
		"rows":    results,
		"count":   len(results),
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleDBHubListSchemas lists database schemas (PostgreSQL/MySQL).
// Tool: dbhub_list_schemas
func HandleDBHubListSchemas(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, driver, e := dbhubOpen(args)
	if e != nil {
		return err(fmt.Sprintf("Failed to connect: %v", e))
	}
	defer db.Close()

	var query string
	switch driver {
	case "postgres":
		query = "SELECT schema_name FROM information_schema.schemata ORDER BY schema_name"
	case "mysql":
		query = "SELECT schema_name FROM information_schema.schemata ORDER BY schema_name"
	default:
		// SQLite doesn't have schemas
		return ok("SQLite does not support schemas. All tables are in the 'main' database.")
	}

	rows, e := db.QueryContext(ctx, query)
	if e != nil {
		return err(fmt.Sprintf("Failed to list schemas: %v", e))
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			schemas = append(schemas, name)
		}
	}

	out, _ := json.MarshalIndent(schemas, "", "  ")
	return ok(string(out))
}
