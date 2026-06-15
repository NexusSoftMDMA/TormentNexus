package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://kibela.example.com/api/search?q="+query, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to call Kibela API")
}

	defer resp.Body.Close()
	return success(fmt.Sprintf("Search results for %s: status %d", query, resp.StatusCode))
}