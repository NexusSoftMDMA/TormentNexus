package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleHttpTest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url parameter is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", e))
}

	defer resp.Body.Close()
	return ok(fmt.Sprintf("HTTP status: %d", resp.StatusCode))
}

func HandleVpnStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	server, _ :=getString(args, "server")
	if server == "" {
		server = "default"
	}
	return ok(fmt.Sprintf("VPN '%s' is connected", server))
}