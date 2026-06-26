package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleEcho_ruflo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
}

	body := fmt.Sprintf(`{"echo": %q}`, msg)
	resp, e := http.DefaultClient.Post("https://httpbin.org/post", "application/json", nil)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return ok(body)
}

func HandleStatus_ruflo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Ruflo MCP server is running")
}