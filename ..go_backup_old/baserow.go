package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	token, _ :=getString(args, "token")
	if baseURL == "" || token == "" {
		return err("base_url and token are required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/database/tables/", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Header.Set("Authorization", "Token "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}
	var tables []map[string]interface{}
	if e := json.Unmarshal(body, &tables); e != nil {
		return err("failed to parse response: " + e.Error())
	}
	result, e := json.MarshalIndent(tables, "", "  ")
	if e != nil {
		return err("failed to marshal result: " + e.Error())
	}
	return ok("Tables: " + string(result))
}

func HandleListRows(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	token, _ :=getString(args, "token")
	tableID, _ :=getInt(args, "table_id")
	if baseURL == "" || token == "" || tableID == 0 {
		return err("base_url, token, and table_id are required")
	}
	url := fmt.Sprintf("%s/api/database/rows/table/%d/", baseURL, tableID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Header.Set("Authorization", "Token "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
	}
	rows, found := result["rows"]
	if !found {
		return err("response missing 'rows' field")
	}
	rowsJSON, e := json.MarshalIndent(rows, "", "  ")
	if e != nil {
		return err("failed to marshal rows: " + e.Error())
	}
	return ok("Rows: " + string(rowsJSON))
}