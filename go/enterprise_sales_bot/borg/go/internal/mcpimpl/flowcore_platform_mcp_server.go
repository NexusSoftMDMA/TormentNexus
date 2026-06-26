package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleListTenants(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, string(body)))
}

	return success(fmt.Sprintf("Tenants: %s", string(body)))
}

func HandleHealthCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("healthy")
}