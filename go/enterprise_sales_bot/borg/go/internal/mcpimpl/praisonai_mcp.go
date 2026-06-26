package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandlePing_praisonai_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		url = "https://example.com"
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to reach URL: " + e.Error())
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return success("Status: " + resp.Status)
}