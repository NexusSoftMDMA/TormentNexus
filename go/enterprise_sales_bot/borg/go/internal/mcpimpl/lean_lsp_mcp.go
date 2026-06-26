package mcpimpl

import (
	"context"
	"net/http"
	"os"
)

func HandleLeanLspMcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := os.Getenv("LEAN_LSP_URL")
	if url == "" {
		url = "http://localhost:7474"
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("connection failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("server returned status: " + resp.Status)
}

	return success("Lean LSP server is running")
}