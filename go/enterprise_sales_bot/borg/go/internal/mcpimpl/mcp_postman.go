package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetCollections_mcp_postman(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.getpostman.com/collections", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("X-Api-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch collections: " + e.Error())
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

	collections, found := result["collections"]
	if !found {
		return err("no collections found")
}

	return ok(fmt.Sprintf("Found %d collections", len(collections.([]interface{}))))
}