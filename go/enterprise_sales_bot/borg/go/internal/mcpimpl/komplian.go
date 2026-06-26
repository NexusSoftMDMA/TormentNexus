package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleSetup(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	resp, e := http.DefaultClient.Get("https://example.com/setup?name=" + name)
	if e != nil {
		return err("HTTP request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok("setup result: " + string(body))
}

func HandleOnboard(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	user, _ :=getString(args, "user")
	if user == "" {
		return err("user is required")
}

	return ok("onboarding started for " + user)
}