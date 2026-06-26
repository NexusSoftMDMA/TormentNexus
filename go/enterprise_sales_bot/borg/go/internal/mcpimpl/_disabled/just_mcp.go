package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleEcho_just_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}

func HandleStatus_just_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://httpbin.org/status/200")
	if e != nil {
		return err("failed to check status")
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("service unavailable")
}

	return success("service is healthy")
}