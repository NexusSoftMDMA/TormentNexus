//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	host, _ :=getString(args, "host")
	user, _ :=getString(args, "user")
	pass, _ :=getString(args, "password")
	db, _ :=getString(args, "database")
	if host == "" || user == "" || pass == "" {
		return err("host, user, and password are required")
}

	url := fmt.Sprintf("https://%s:443/api/v2/query?database=%s", host, db)
	body := fmt.Sprintf(`{"query":"%s"}`, strings.ReplaceAll(query, `"`, `\"`))
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.SetBasicAuth(user, pass)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(data)))
}

	var result interface{	if e := json.Unmarshal(data, &result); e != nil {
		return err(fmt.Sprintf("invalid JSON response: %v", e))
}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(jsonBytes))
}

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	user, _ :=getString(args, "user")
	pass, _ :=getString(args, "password")
	db, _ :=getString(args, "database")
	if host == "" || user == "" || pass == "" {
		return err("host, user, and password are required")
}

	url := fmt.Sprintf("https://%s:443/api/v2/query?database=%s", host, db)
	body := `{"query":"SHOW TABLES"}`
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.SetBasicAuth(user, pass)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(data)))
}

	var result interface{	if e := json.Unmarshal(data, &result); e != nil {
		return err(fmt.Sprintf("invalid JSON response: %v", e))
}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(jsonBytes))
}
}
}
}
}