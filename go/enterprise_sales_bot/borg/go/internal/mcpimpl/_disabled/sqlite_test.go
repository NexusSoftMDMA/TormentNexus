package mcpimpl

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestHandleSqliteTools(t *testing.T) {
	// Create a temp directory for DB
	tempDir, err := os.MkdirTemp("", "sqlite-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, errDb := sql.Open("sqlite", dbPath)
	if errDb != nil {
		t.Fatalf("Failed to create test DB: %v", errDb)
	}

	// Create test table
	_, errCreate := db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)")
	if errCreate != nil {
		db.Close()
		t.Fatalf("Failed to create table: %v", errCreate)
	}

	// Insert row
	_, errInsert := db.Exec("INSERT INTO users (name, email) VALUES ('Alice', 'alice@example.com')")
	if errInsert != nil {
		db.Close()
		t.Fatalf("Failed to insert row: %v", errInsert)
	}
	db.Close()

	// 1. Test HandleSqliteGetCatalog
	argsCatalog := map[string]interface{}{
		"sqlite_file": dbPath,
	}
	respCatalog, errCat := HandleSqliteGetCatalog(context.Background(), argsCatalog)
	if errCat != nil {
		t.Fatalf("GetCatalog returned error: %v", errCat)
	}
	if respCatalog.IsError {
		t.Fatalf("GetCatalog response contains error: %s", respCatalog.Content[0].Text)
	}

	catalogText := respCatalog.Content[0].Text
	if !strings.Contains(catalogText, "users") || !strings.Contains(catalogText, "email") {
		t.Errorf("Expected 'users' table and 'email' column in catalog, got: %s", catalogText)
	}

	// 2. Test HandleSqliteExecute Query (Read-Only)
	argsQuery := map[string]interface{}{
		"sqlite_file": dbPath,
		"sql":         "SELECT name, email FROM users WHERE id = 1",
	}
	respQuery, errQ := HandleSqliteExecute(context.Background(), argsQuery)
	if errQ != nil {
		t.Fatalf("Execute query returned error: %v", errQ)
	}
	if respQuery.IsError {
		t.Fatalf("Execute query response contains error: %s", respQuery.Content[0].Text)
	}

	queryText := respQuery.Content[0].Text
	if !strings.Contains(queryText, "<td>Alice</td>") || !strings.Contains(queryText, "<th>email</th>") {
		t.Errorf("Expected HTML table with Alice and email header, got: %s", queryText)
	}

	// 3. Test HandleSqliteExecute Statement (Write)
	argsWrite := map[string]interface{}{
		"sqlite_file": dbPath,
		"sql":         "INSERT INTO users (name, email) VALUES ('Bob', 'bob@example.com')",
	}
	respWrite, errW := HandleSqliteExecute(context.Background(), argsWrite)
	if errW != nil {
		t.Fatalf("Execute write returned error: %v", errW)
	}
	if respWrite.IsError {
		t.Fatalf("Execute write response contains error: %s", respWrite.Content[0].Text)
	}

	writeText := respWrite.Content[0].Text
	if !strings.Contains(writeText, "Statement executed successfully") {
		t.Errorf("Expected success message, got: %s", writeText)
	}
}
