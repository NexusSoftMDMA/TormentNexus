//go:build ignore
// +build ignore

package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"strings"

	_ "modernc.org/sqlite"

)

// TableCatalog holds metadata about columns
type TableCatalog struct {
	Columns map[string]string `json:"columns"`
}

// DatabaseCatalog holds metadata about tables
type DatabaseCatalog struct {
	Title  string                  `json:"title"`
	Tables map[string]TableCatalog `json:"tables"`
}

// RootCatalog maps database name to database metadata
type RootCatalog struct {
	Databases map[string]DatabaseCatalog `json:"databases"`
}

// HandleSqliteGetCatalog returns a JSON summary of databases, tables, and columns.
func HandleSqliteGetCatalog(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dbPath, _ := getString(args, "db_path", "dbPath", "sqlite_file")
	if dbPath == "" {
		dbPath = "tormentnexus.db" // default fallback
	}

	// Open read-only
	db, errDb := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=ro", dbPath))
	if errDb != nil {
		return err(fmt.Sprintf("Failed to open database: %v", errDb))
	}
	defer db.Close()

	catalog := RootCatalog{
		Databases: make(map[string]DatabaseCatalog),
	}

	// Query database list
	dbRows, errQuery := db.QueryContext(ctx, "PRAGMA database_list")
	if errQuery != nil {
		return err(fmt.Sprintf("Failed to query database list: %v", errQuery))
	}
	defer dbRows.Close()

	type dbInfo struct {
		seq  int
		name string
		file string
	}
	var dbs []dbInfo
	for dbRows.Next() {
		var d dbInfo
		if errScan := dbRows.Scan(&d.seq, &d.name, &d.file); errScan != nil {
			continue
		}
		dbs = append(dbs, d)
	}

	for _, d := range dbs {
		dbCatalog := DatabaseCatalog{
			Title:  d.name,
			Tables: make(map[string]TableCatalog),
		}

		// Query table list from sqlite_master to avoid sqlite_ internal tables
		tableQuery := fmt.Sprintf("SELECT name FROM %s.sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%%'", d.name)
		tableRows, errTable := db.QueryContext(ctx, tableQuery)
		if errTable != nil {
			// fallback if sqlite_master query fails
			continue
		}

		var tableNames []string
		for tableRows.Next() {
			var name string
			if errScan := tableRows.Scan(&name); errScan == nil {
				tableNames = append(tableNames, name)
			}
		}
		tableRows.Close()

		for _, tableName := range tableNames {
			tableCatalog := TableCatalog{
				Columns: make(map[string]string),
			}

			// Query column info
			colQuery := fmt.Sprintf("PRAGMA %s.table_info(%s)", d.name, tableName)
			colRows, errCol := db.QueryContext(ctx, colQuery)
			if errCol != nil {
				continue
			}

			for colRows.Next() {
				var cid int
				var name, colType string
				var notnull int
				var dfltVal interface{}
				var pk int
				if errScan := colRows.Scan(&cid, &name, &colType, &notnull, &dfltVal, &pk); errScan == nil {
					tableCatalog.Columns[name] = colType
				}
			}
			colRows.Close()

			dbCatalog.Tables[tableName] = tableCatalog
		}

		catalog.Databases[d.name] = dbCatalog
	}

	jsonData, errMarshal := json.MarshalIndent(catalog, "", "  ")
	if errMarshal != nil {
		return err(fmt.Sprintf("Failed to marshal catalog JSON: %v", errMarshal))
	}

	return ok(string(jsonData))
}

// HandleSqliteExecute runs a raw SQL query and returns an HTML-formatted table of the output.
func HandleSqliteExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sqlQuery, _ := getString(args, "sql", "query")
	if sqlQuery == "" {
		return err("sql parameter is required")
	}

	dbPath, _ := getString(args, "db_path", "dbPath", "sqlite_file")
	if dbPath == "" {
		dbPath = "tormentnexus.db" // default fallback
	}

	// Determine read-only vs read-write based on query string
	mode := "ro"
	isWrite := false
	upperSQL := strings.ToUpper(strings.TrimSpace(sqlQuery))
	if strings.HasPrefix(upperSQL, "INSERT") || strings.HasPrefix(upperSQL, "UPDATE") || strings.HasPrefix(upperSQL, "DELETE") || strings.HasPrefix(upperSQL, "CREATE") || strings.HasPrefix(upperSQL, "DROP") {
		mode = "rw"
		isWrite = true
	}

	db, errDb := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=%s", dbPath, mode))
	if errDb != nil {
		return err(fmt.Sprintf("Failed to open database: %v", errDb))
	}
	defer db.Close()

	if isWrite {
		res, errExec := db.ExecContext(ctx, sqlQuery)
		if errExec != nil {
			return err(fmt.Sprintf("Failed to execute SQL statement: %v", errExec))
		}
		rowsAffected, _ := res.RowsAffected()
		return ok(fmt.Sprintf("Statement executed successfully. Rows affected: %d", rowsAffected))
	}

	rows, errQuery := db.QueryContext(ctx, sqlQuery)
	if errQuery != nil {
		return err(fmt.Sprintf("Failed to execute SQL query: %v", errQuery))
	}
	defer rows.Close()

	cols, errCols := rows.Columns()
	if errCols != nil {
		return err(fmt.Sprintf("Failed to retrieve columns: %v", errCols))
	}

	var headerBuilder strings.Builder
	for _, col := range cols {
		headerBuilder.WriteString(fmt.Sprintf("<th>%s</th>", html.EscapeString(col)))
	}
	rowsHtml := fmt.Sprintf("<tr>%s</tr>", headerBuilder.String())

	// Read rows
	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if errScan := rows.Scan(valuePtrs...); errScan != nil {
			continue
		}

		var rowBuilder strings.Builder
		for _, val := range values {
			valStr := "NULL"
			if val != nil {
				switch v := val.(type) {
				case []byte:
					valStr = string(v)
				default:
					valStr = fmt.Sprintf("%v", v)
				}
			}
			rowBuilder.WriteString(fmt.Sprintf("<td>%s</td>", html.EscapeString(valStr)))
		}
		rowsHtml += fmt.Sprintf("<tr>%s</tr>", rowBuilder.String())
	}

	return ok(fmt.Sprintf("<table>%s</table>", rowsHtml))
}
