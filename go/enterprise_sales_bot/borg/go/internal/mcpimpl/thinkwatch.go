package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func HandleAuditLog(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	user, _ :=getString(args, "user")
	action, _ :=getString(args, "action")
	resource, _ :=getString(args, "resource")
	if user == "" || action == "" {
		return err("missing user or action")
}

	entry := map[string]interface{}{
		"timestamp": time.Now().UTC(),
		"user":      user,
		"action":    action,
		"resource":  resource,
		"status":    "allowed",
	}
	data, e := json.Marshal(entry)
	if e != nil {
		return err("failed to marshal log")
}

	return success(string(data))
}

func HandleRateCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	user, _ :=getString(args, "user")
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 100
	}
	if user == "" {
		return err("missing user for rate check")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/rate/"+user, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("rate service unavailable")
}

	defer resp.Body.Close()
	return ok("rate limit checked")
}