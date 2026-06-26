package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleExecuteSQL_mcp_server_neon(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	if sql == "" {
		return err("sql is required")
}

	apiURL, _ :=getString(args, "neon_api_url")
	apiKey, _ :=getString(args, "neon_api_key")
	if apiURL == "" || apiKey == "" {
		return err("neon_api_url and neon_api_key are required")
}

	body := fmt.Sprintf(`{"query":"%s"}`, sql)
	req, e := http.NewRequestWithContext(ctx, "POST", apiURL+"/sql", strings.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API returned " + resp.Status + ": " + string(respBody))
}

	var result interface{}
	json.Unmarshal(respBody, &result)
	out, _ := json.Marshal(result)
	return ok(string(out))
}

func HandleListTables_mcp_server_neon(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	newArgs := map[string]interface{}{
		"sql":           "SELECT table_name FROM information_schema.tables WHERE table_schema='public'",
		"neon_api_url":  getString(args, "neon_api_url"),
		"neon_api_key":  getString(args, "neon_api_key"),
	}
	return HandleExecuteSQL(ctx, newArgs)
}