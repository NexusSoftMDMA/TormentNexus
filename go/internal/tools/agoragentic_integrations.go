//go:build ignore
// +build ignore

package tools

import (
	"context"
	"net/http"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		message = "Agoragentic Integrations MCP server is running"
	}
	return success(message)
}

func HandleHealth(_ context.Context, _ map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://example.com/health")
	if e != nil {
		return err("health check failed: " + e.Error())
}

	resp.Body.Close()
	return ok("healthy")
}
