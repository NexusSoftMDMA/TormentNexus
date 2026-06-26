package tools

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleReadNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("missing path")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "http://localhost:27123/vault/"+url.PathEscape(path), nil)
	if e != nil {
		return err("request failed: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	return ok(string(body))
}

func HandleSearchNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing query")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "http://localhost:27123/search/"+url.PathEscape(query), nil)
	if e != nil {
		return err("request failed: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	return ok(string(body))
}