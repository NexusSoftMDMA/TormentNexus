package tools

import (
	"context"
	"net/http"
)

func HandleInboxZero(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter required")
}

	resp, e := http.DefaultClient.Get("https://example.com?q=" + query)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	return success("Inbox Zero query sent: " + query)
}