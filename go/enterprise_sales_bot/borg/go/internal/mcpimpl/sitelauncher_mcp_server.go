package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetSite_sitelauncher_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	io.ReadAll(resp.Body)
	return success(fmt.Sprintf("Fetched %s (status %d)", url, resp.StatusCode))
}

func HandleListSites_sitelauncher_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sites := []string{"https://example.com", "https://github.com"}
	return ok(fmt.Sprintf("Available sites: %v", sites))
}