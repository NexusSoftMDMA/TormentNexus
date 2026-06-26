package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListMcpServers_mcp_servers_live(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.github.com/search/repositories?q=topic:anthropic-mcp")
	if e != nil {
		return err("failed to fetch servers: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result struct {
		Items []struct {
			FullName string `json:"full_name"`
		} `json:"items"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse: " + e.Error())
}

	names := ""
	for _, item := range result.Items {
		names += item.FullName + "\n"
	}
	return ok("Servers:\n" + names)
}

func HandleGetMcpServer_mcp_servers_live(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("'name' is required")
}

	resp, e := http.DefaultClient.Get("https://api.github.com/search/repositories?q=" + name + "+topic:anthropic-mcp")
	if e != nil {
		return err("failed to fetch server: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result struct {
		Items []struct {
			Description string `json:"description"`
		} `json:"items"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse: " + e.Error())
}

	if len(result.Items) == 0 {
		return err("server not found")
}

	return success("Description: " + result.Items[0].Description)
}