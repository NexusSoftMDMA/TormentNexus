package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HandleGetProject fetches a project by ID from an external API.
func HandleGetProject_builders_sodax_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "project_id")
	resp, e := http.DefaultClient.Get("https://api.example.com/projects/" + projectID)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch project: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	// Optionally parse and format
	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("invalid JSON: %v", e))
}

	result, _ := json.MarshalIndent(data, "", "  ")
	return ok(string(result))
}