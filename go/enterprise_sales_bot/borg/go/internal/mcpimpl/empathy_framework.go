package mcpimpl

import (
	"context"
	"fmt"
	"net/http"
)

func HandleEmpathy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	_, e := http.DefaultClient.Get("http://localhost:9999/empathy?q=" + message)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	return success("empathy processed for: " + message)
}