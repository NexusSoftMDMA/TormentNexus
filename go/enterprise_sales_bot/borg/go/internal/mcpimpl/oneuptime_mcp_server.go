package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListIncidents_oneuptime_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "projectId")
	if projectID == "" {
		return err("projectId is required")
}

	baseURL := os.Getenv("ONEUPTIME_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.oneuptime.com"
	}
	apiKey := os.Getenv("ONEUPTIME_API_KEY")
	if apiKey == "" {
		return err("ONEUPTIME_API_KEY environment variable not set")
}

	url := fmt.Sprintf("%s/incidents?projectId=%s", baseURL, projectID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("API returned status " + fmt.Sprintf("%d", resp.StatusCode) + ": " + string(body))
}

	return ok(string(body))
}