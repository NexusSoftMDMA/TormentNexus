package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleListIssues(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	teamID, _ :=getString(args, "teamId")
	if teamID == "" {
		return err("teamId is required")
	}
	token := os.Getenv("LINEAR_API_TOKEN")
	if token == "" {
		return err("LINEAR_API_TOKEN not set")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.linear.app/rest/v1/issues?teamId="+teamID, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
	}
	out := fmt.Sprintf("Found %d issues", len(result.Data))
	return success(out)
}