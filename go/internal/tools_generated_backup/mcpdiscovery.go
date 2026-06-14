package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://raw.githubusercontent.com/awesome-mcp-servers/main/README.md")
	if e != nil {
		return err("failed to fetch MCP servers list")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}