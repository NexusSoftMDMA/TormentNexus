package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleListPods_kubernetes_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://localhost:8080/api/v1/pods")
	if e != nil {
		return err("failed to list pods: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API returned status " + resp.Status)
}

	return ok(string(body))
}

func HandleGetNode_kubernetes_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	nodeName, _ :=getString(args, "name")
	if nodeName == "" {
		return err("node name is required")
}

	url := "http://localhost:8080/api/v1/nodes/" + nodeName
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get node: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API returned status " + resp.Status)
}

	return ok(string(body))
}