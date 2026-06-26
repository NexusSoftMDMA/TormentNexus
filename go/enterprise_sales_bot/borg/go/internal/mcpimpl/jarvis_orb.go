package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleJarvisOrb(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	resp, e := http.DefaultClient.Get("https://httpbin.org/get?q=" + query)
	if e != nil {
		return err("HTTP request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}