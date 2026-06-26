package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleQuery_mcp_snowflake_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	body, _ := json.Marshal(map[string]string{"statement": query})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://snowflake.example.com/api/v2/statements", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Query result: %v", result))
}

func HandleListTables_mcp_snowflake_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	db, _ :=getString(args, "database")
	schema, _ :=getString(args, "schema")
	query := fmt.Sprintf("SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_CATALOG='%s' AND TABLE_SCHEMA='%s'", db, schema)
	body, _ := json.Marshal(map[string]string{"statement": query})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://snowflake.example.com/api/v2/statements", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	if data, found := result["data"]; found {
		return ok(fmt.Sprintf("Tables: %v", data))
}

	return ok("No tables found")
}