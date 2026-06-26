package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetProxyStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	proxyID, _ :=getString(args, "proxy_id")
	if proxyID == "" {
		return err("proxy_id is required")
}

	url := fmt.Sprintf("https://api.lightningprox.example.com/status/%s", proxyID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get status: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Proxy status: %v", data))
}

func HandleListProxies_lightningprox_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.lightningprox.example.com/proxies"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to list proxies: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var proxies []interface{}
	if e := json.Unmarshal(body, &proxies); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return success(fmt.Sprintf("Found %d proxies", len(proxies)))
}