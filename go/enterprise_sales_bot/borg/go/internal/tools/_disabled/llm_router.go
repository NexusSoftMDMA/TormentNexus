package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleListModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	models := []string{"gpt-4", "gpt-3.5", "claude-v2"}
	data, e := json.Marshal(models)
	if e != nil {
		return err("failed to marshal models")
	}
	return ok(string(data))
}

func HandleChat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
	}
	resp, e := http.DefaultClient.Post("https://api.example.com/chat", "application/json", nil)
	if e != nil {
		return err("API call failed: " + e.Error())
	}
	defer resp.Body.Close()
	return ok("Echo: " + message)
}