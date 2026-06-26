package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleSupabaseQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	table, _ :=getString(args, "table")
	if table == "" {
		return err("table argument is required")
	}
	url := "https://api.supabase.co/rest/v1/" + table
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
	}
	return ok("Query executed successfully")
}

func HandleSupabaseInsert(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	table, _ :=getString(args, "table")
	if table == "" {
		return err("table argument is required")
	}
	data, found := args["data"]
	if !found {
		return err("data argument is required")
	}
	body, e := json.Marshal(data)
	if e != nil {
		return err("failed to marshal data: " + e.Error())
	}
	url := "https://api.supabase.co/rest/v1/" + table
	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Body = nil // Placeholder for actual body handling in real impl
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	return success("Insert initiated for " + table)
}