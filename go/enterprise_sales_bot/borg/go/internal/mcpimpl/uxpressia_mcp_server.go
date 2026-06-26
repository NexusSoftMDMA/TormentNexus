package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleGetProject_uxpressia_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "projectId")
	if id == "" {
		return err("missing projectId")
}

	url := "https://api.uxpressia.com/v1/projects/" + id
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}

func HandleListProjects_uxpressia_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.uxpressia.com/v1/projects"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}