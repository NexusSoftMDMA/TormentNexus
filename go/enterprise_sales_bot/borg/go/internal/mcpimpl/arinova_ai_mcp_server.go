package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleAction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	if action == "" {
		return err("missing action")
	}
	base := os.Getenv("ARINOVA_BASE_URL")
	if base == "" {
		base = "http://localhost:8080"
	}
	url := fmt.Sprintf("%s/actions/%s", base, action)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: "+e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: "+e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response body: "+e.Error())
	}
	return ok(string(body))
}

func HandleStatus_arinova_ai_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := os.Getenv("ARINOVA_BASE_URL")
	if base == "" {
		base = "http://localhost:8080"
	}
	resp, e := http.DefaultClient.Get(base + "/health")
	if e != nil {
		return err("health check failed: "+e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("unhealthy: "+resp.Status)
	}
	return ok("healthy")
}