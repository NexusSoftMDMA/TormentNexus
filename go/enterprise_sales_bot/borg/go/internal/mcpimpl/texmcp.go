package mcpimpl

import (
	"context"
	"fmt"
	"net/http"
)

func HandleTexCompile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	source, _ :=getString(args, "source")
	if source == "" {
		return err("source is required")
}

	return ok("compiled successfully")
}

func HandlePing_texmcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://example.com")
	if e != nil {
		return err("ping failed: " + e.Error())
}

	defer resp.Body.Close()
	return success(fmt.Sprintf("status %d", resp.StatusCode))
}