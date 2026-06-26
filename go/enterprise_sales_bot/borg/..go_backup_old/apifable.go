package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleQueryTable(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	table, _ :=getString(args, "table")
	query, _ :=getString(args, "query")
	url := fmt.Sprintf("https://api.apifable.com/v1/tables/%s?q=%s", table, query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return success(fmt.Sprintf("Result: %v", result))
}

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.apifable.com/v1/tables")
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var tables []string
	if e := json.NewDecoder(resp.Body).Decode(&tables); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return success(fmt.Sprintf("Tables: %v", tables))
}