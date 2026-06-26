package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// HandleAskAgent sends a message to a Maiat agent and returns the response.
func HandleAskAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	agentID, _ :=getString(args, "agent_id")
	message, _ :=getString(args, "message")
	if agentID == "" || message == "" {
		return err("agent_id and message are required")
}

	u := fmt.Sprintf("https://api.maiat.ai/v1/agents/%s/ask?message=%s",
		url.PathEscape(agentID), url.QueryEscape(message))

	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return ok(string(body))
}

	if msg, found := result["response"]; found {
		return ok(fmt.Sprint(msg))
}

	return ok(string(body))
}