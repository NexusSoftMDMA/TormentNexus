package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	docsURL, _ :=getString(args, "docs_url")
	query, _ :=getString(args, "query")
	if docsURL == "" || query == "" {
		return err("docs_url and query are required")
}

	searchURL := fmt.Sprintf("%s/search?q=%s", docsURL, url.QueryEscape(query))
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}