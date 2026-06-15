//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func HandleGetScore(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	agentID, _ :=getString(args, "agent_id")
	url := fmt.Sprintf("https://api.agentscore.example.com/score/%s", agentID)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch score")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("invalid response")
}

	score, found := result["score"].(float64)
	if !found {
		return err("score not found")
}

	return ok(fmt.Sprintf("Agent %s score: %.2f", agentID, score))
}

func HandleListAgents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.agentscore.example.com/agents", nil)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to list agents")
}

	defer resp.Body.Close()
	var agents []string
	if e = json.NewDecoder(resp.Body).Decode(&agents); e != nil {
		return err("invalid response")
}

	data, _ := json.Marshal(agents)
	return success(string(data))
}
