package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleRunAction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "server_url")
	if base == "" {
		return err("server_url is required")
}

	agentId, _ :=getString(args, "agent_id")
	if agentId == "" {
		return err("agent_id is required")
}

	action, _ :=getString(args, "action")
	if action == "" {
		return err("action is required")
}

	url := strings.TrimRight(base, "/") + "/agents/" + agentId + "/actions"
	bodyMap := map[string]interface{}{
		"action": action,
	}
	if params, found := args["params"]; found {
		bodyMap["params"] = params
	}
	body, e := json.Marshal(bodyMap)
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(body)))
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http error: %v", e))
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read error: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(respBody)))
}

	return ok(string(respBody))
}

func HandleGetAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "server_url")
	if base == "" {
		return err("server_url is required")
}

	agentId, _ :=getString(args, "agent_id")
	if agentId == "" {
		return err("agent_id is required")
}

	url := strings.TrimRight(base, "/") + "/agents/" + agentId
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http error: %v", e))
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read error: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(respBody)))
}

	return ok(string(respBody))
}