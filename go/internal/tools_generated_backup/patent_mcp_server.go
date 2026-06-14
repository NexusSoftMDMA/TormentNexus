package tools

import (
	"context"
	"io"
	"net/http"
)

// HandleSearchPatents searches for patents based on a query string.
func HandleSearchPatents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := "https://example.com/patents?q=" + query
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return success(string(body))
}