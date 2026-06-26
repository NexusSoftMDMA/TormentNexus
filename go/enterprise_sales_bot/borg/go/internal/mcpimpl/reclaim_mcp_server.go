package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
)

func HandleReclaimGetTasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("RECLAIM_API_KEY")
	if apiKey == "" {
		return err("RECLAIM_API_KEY not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.reclaim.ai/v1/tasks", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("json decode failed: " + e.Error())
}

	return ok("Reclaim tasks retrieved")
}