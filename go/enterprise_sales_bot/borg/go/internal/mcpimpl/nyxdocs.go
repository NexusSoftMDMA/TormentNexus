package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleListDocs_nyxdocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	docs := []string{"overview", "getting-started", "api-reference"}
	return success(fmt.Sprintf("Available docs: %v", docs))
}

func HandleGetDoc_nyxdocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url parameter is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch doc: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return success(string(body))
}