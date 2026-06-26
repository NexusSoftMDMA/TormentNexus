package mcpimpl

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleExecute_claude_terminal_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	apiURL := "http://localhost:9090/terminal?command=" + url.QueryEscape(cmd)
	resp, e := http.DefaultClient.Get(apiURL)
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