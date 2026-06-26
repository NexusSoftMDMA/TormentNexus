package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleSearch_glean_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	token := os.Getenv("GLEAN_API_TOKEN")
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.glean.com/rest/api/v1/search", nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	body, _ := json.Marshal(map[string]string{"query": query})
	req.Body = http.NoBody
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed")
}

	summary := fmt.Sprintf("Search returned %d results", len(result))
	return ok(summary)
}

func HandleGetDocument_glean_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	token := os.Getenv("GLEAN_API_TOKEN")
	url := "https://api.glean.com/rest/api/v1/document/" + id
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	var doc map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&doc); e != nil {
		return err("decode failed")
}

	title, _ := doc["title"].(string)
	return ok("Document: " + title)
}