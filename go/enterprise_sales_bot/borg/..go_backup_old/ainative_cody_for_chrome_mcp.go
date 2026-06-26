package tools

import (
	"context"
	"net/http"
)

func HandleFetchPage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return ok("no url provided")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("fetch failed: " + e.Error())
}

	defer resp.Body.Close()
	return success("page fetched, status: " + resp.Status)
}

func HandleQueryCody(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return ok("no query provided")
}

	return success("cody query processed: " + query)
}