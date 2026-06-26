package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetProxyInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "proxy_id")
	if id == "" {
		return err("proxy_id is required")
}

	url := fmt.Sprintf("https://demcp.io/api/proxy/%s/info", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch proxy info: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleListProxies_awesome_web3_mcp_servers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	url := fmt.Sprintf("https://demcp.io/api/proxies?limit=%d", limit)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to list proxies: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}