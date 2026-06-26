package tools

import (
	"context"
	"net/http"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	resp, e := http.DefaultClient.Get("https://example.com?q=" + query)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	return success("fetched " + query)
}

func HandleY(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	value, _ :=getInt(args, "count")
	if value > 100 {
		return err("count too large")
}

	return ok("count accepted")
}