package tools

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleRenderMermaid(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	resp, e := http.DefaultClient.Get("https://mermaid.ink/svg/" + url.QueryEscape(code))
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}