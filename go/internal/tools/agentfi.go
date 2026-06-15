//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListAgents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.agentfi.io/agents", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("bad status: %d", resp.StatusCode))
}

	var result struct {
		Agents []interface{} `json:"agents"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return success(result)
}

func HandleGetAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	agentID, _ :=getString(args, "agentID")
	url := fmt.Sprintf("https://api.agentfi.io/agents/%s", agentID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("status %d", resp.StatusCode))
}

	var agent interface{}
	if e := json.NewDecoder(resp.Body).Decode(&agent); e != nil {
		return err("decode: " + e.Error())
}

	return success(agent)
}
