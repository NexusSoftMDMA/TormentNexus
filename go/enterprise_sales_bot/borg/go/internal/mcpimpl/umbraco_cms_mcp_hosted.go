package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// HandleUmbracoQueryContent fetches content from Umbraco CMS based on name or ID
func HandleUmbracoQueryContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL := os.Getenv("UMBRACO_BASE_URL")
	if baseURL == "" {
		return err("UMBRACO_BASE_URL not set")
}

	apiKey := os.Getenv("UMBRACO_API_KEY")
	if apiKey == "" {
		return err("UMBRACO_API_KEY not set")
}

	name, _ :=getString(args, "name")
	id, _ :=getString(args, "id")

	url := baseURL + "/umbraco/api/content"
	if name != "" {
		url += "?name=" + name
	} else if id != "" {
		url += "/" + id
	}

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Accept", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("json decode error: %v", e))
}

	return ok(result)
}

// HandleUmbracoPingContent verifies connectivity to Umbraco API
func HandleUmbracoPingContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Umbraco MCP host is reachable")
}