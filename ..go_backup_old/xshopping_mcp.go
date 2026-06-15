package tools

import (
	"context"
	"net/http"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	query, _ :=getString(args, "query")

	if action != "search" {
		return err("unsupported action: " + action)
}

	if query == "" {
		return err("query is required")
}

	resp, e := http.DefaultClient.Get("https://api.example.com/search?q=" + query)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	return success("search results for: " + query)
}