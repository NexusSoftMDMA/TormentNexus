package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleListUsers_webmin_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "http://localhost:10000/session_list.cgi?type=users", nil)
	if e != nil {
		return err("create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response: " + e.Error())
}

	return success(string(body))
}

func HandleRestartService(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	service, _ :=getString(args, "service")
	url := "http://localhost:10000/service/restart.cgi?service=" + service
	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err("create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response: " + e.Error())
}

	return success(string(body))
}