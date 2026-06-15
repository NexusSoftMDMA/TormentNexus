//go:build ignore
// +build ignore

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListAgents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fleetID, _ :=getString(args, "fleet_id")
	url := "https://api.agent-fleet-o.example.com/agents"
	if fleetID != "" {
		url += "?fleet_id=" + fleetID
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return success(fmt.Sprintf("Agents: %v", result))
}

func HandleDispatchAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	agentID, _ :=getString(args, "agent_id")
	command, _ :=getString(args, "command")
	if agentID == "" || command == "" {
		return err("both agent_id and command are required")
}

	payload := map[string]string{"agent_id": agentID, "command": command}
	body, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal payload: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.agent-fleet-o.example.com/dispatch", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return err("dispatch failed with status " + resp.Status)
}

	return success("Agent dispatched")
}
