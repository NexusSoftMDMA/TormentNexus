package mcpimpl

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleQueryDocumentation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	base := "https://api.macroforge.ai/docs"
	reqURL, e := url.Parse(base + "?q=" + url.QueryEscape(query))
	if e != nil {
		return err("invalid URL: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	res, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer res.Body.Close()
	body, e := io.ReadAll(res.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}