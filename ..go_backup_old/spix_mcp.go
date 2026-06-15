package tools

import (
	"context"
	"net/http"
	"io"
)

// HandleSpixInfo returns info from Spix Mcp server
func HandleSpixInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://spix-mcp.example.com/info"
	if v := getString(args, "url"); v != "" {
		url = v
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}