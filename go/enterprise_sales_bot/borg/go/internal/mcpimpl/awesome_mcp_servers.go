package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleGetAwesomeMcpServers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://raw.githubusercontent.com/punkpeye/awesome-mcp-servers/main/README.md")
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read: " + e.Error())
}

	return ok(string(body))
}