package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleQuery_mcp_server_arangodb(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url, _ :=getString(args, "url")
	if url == "" {
		url = "http://localhost:8529"
	}
	database, _ :=getString(args, "database")
	if database == "" {
		database = "_system"
	}
	user, _ :=getString(args, "username")
	pass, _ :=getString(args, "password")

	body := map[string]interface{}{"query": query}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", url+"/_db/"+database+"/_api/cursor", strings.NewReader(string(jsonBody)))
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	if user != "" || pass != "" {
		req.SetBasicAuth(user, pass)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read error: %v", e))
}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(respBody)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err(fmt.Sprintf("JSON error: %v", e))
}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}
}