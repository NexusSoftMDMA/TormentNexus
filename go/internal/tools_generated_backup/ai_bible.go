package tools

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchBible(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query required")
}

	resp, e := http.DefaultClient.Get("https://bible-api.com?query=" + url.QueryEscape(query))
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