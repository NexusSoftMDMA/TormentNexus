package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	if sql == "" {
		return err("sql is required")
}

	url := os.Getenv("COCKROACHDB_HTTP_URL")
	if url == "" {
		return err("COCKROACHDB_HTTP_URL not set")
}

	body := fmt.Sprintf(`{"database":"defaultdb","sql":"%s"}`, strings.ReplaceAll(sql, "\"", "\\\""))
	req, e := http.NewRequestWithContext(ctx, "POST", url+"/api/v1/sql", strings.NewReader(body))
	if e != nil {
		return err("request creation: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http request: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode: " + e.Error())
}

	return ok(fmt.Sprintf("Query result: %v", result))
}

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, _ :=getString(args, "database")
	if db == "" {
		db = "defaultdb"
	}
	url := os.Getenv("COCKROACHDB_HTTP_URL")
	if url == "" {
		return err("COCKROACHDB_HTTP_URL not set")
}

	sql := fmt.Sprintf("SELECT table_name FROM information_schema.tables WHERE table_schema = '%s'", db)
	body := fmt.Sprintf(`{"database":"%s","sql":"%s"}`, db, strings.ReplaceAll(sql, "\"", "\\\""))
	req, e := http.NewRequestWithContext(ctx, "POST", url+"/api/v1/sql", strings.NewReader(body))
	if e != nil {
		return err("request creation: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http request: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode: " + e.Error())
}

	return ok(fmt.Sprintf("Tables: %v", result))
}