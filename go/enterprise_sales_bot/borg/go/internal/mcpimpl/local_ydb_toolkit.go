package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleExecuteQuery_local_ydb_toolkit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	database, _ :=getString(args, "database")
	if query == "" || database == "" {
		return err("query and database are required")
}

	body, _ := json.Marshal(map[string]string{"query": query})
	req, e := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("http://localhost:8765/yql?database=%s", database), bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("ydb returned status " + resp.Status)
}

	return ok("query executed")
}

func HandleListTables_local_ydb_toolkit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	database, _ :=getString(args, "database")
	if database == "" {
		return err("database is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://localhost:8765/tables?database=%s", database), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return success(fmt.Sprintf("tables: %v", result))
}