package mcpimpl

import (
	"context"
	"net/http"
)

func HandleBeep(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		msg = "Beep!"
	}
	return ok(msg)
}

func HandleBeepStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/beep/status", nil)
	if e != nil {
		return err("failed to create request")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch status")
	}
	defer resp.Body.Close()
	return ok("Beeper status: active")
}