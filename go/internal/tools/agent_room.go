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
	resp, e := http.DefaultClient.Get("https://api.agentroom.com/agents")
	if e != nil {
		return err("failed to fetch agents: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	var agents []string
	if e := json.Unmarshal(body, &agents); e != nil {
		return err("failed to parse agents: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d agents", len(agents)))
}

func HandleSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	payload := map[string]string{"message": message}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	resp, e := http.DefaultClient.Post("https://api.agentroom.com/messages", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to send: " + e.Error())
}

	defer resp.Body.Close()
	return ok("Message sent")
}
