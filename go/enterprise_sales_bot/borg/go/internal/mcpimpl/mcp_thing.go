package mcpimpl

import (
	"context"
	"net/http"
	"time"
)

func HandleEcho_mcp_thing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}

func HandleCurrentTime_mcp_thing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	format, _ :=getString(args, "format")
	if format == "" {
		format = time.RFC3339
	}
	now := time.Now().Format(format)
	return ok(now)
}

func HandleHttpExample(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	found := resp.StatusCode == 200
	if found {
		return success("request succeeded")
}

	return err("request failed")
}