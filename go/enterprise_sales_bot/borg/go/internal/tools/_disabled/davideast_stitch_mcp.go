package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleStitch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	if action == "" {
		return err("action is required")
}

	port, _ :=getInt(args, "port")
	if port <= 0 {
		return err("port must be positive integer")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://localhost:%d/health", port), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to connect: " + e.Error())
}

	resp.Body.Close()
	return ok(fmt.Sprintf("proxy on port %d is active", port))
}