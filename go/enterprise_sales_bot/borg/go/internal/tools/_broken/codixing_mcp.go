package mcpimpl

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchCode_codixing_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	language, _ :=getString(args, "language")
	u := "https://api.codixing.com/search?q=" + url.QueryEscape(query)
	if language != "" {
		u += "&lang=" + url.QueryEscape(language)

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}
}