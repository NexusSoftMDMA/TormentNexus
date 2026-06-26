package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleGreeting(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return ok("Hello, " + name + "! This is Spacebridge MCP.")
}

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	resp, e := http.DefaultClient.Get("https://api.spacebridge.io/query?q=" + query)
	if e != nil {
		return err("API call failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	return ok(string(body))
}