package mcpimpl

import (
	"context"
	"net/http"
	"io"
	"encoding/json"
)

func HandleProcess_bgpt_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
}

	resp, e := http.DefaultClient.Get("https://httpbin.org/post")
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json error: " + e.Error())
}

	return ok("processed: " + msg)
}