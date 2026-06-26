package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListServices_opslevel_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "api_token")
	base, _ :=getString(args, "base_url")
	url := base + "/api/services"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer " + token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Failed to execute request: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleGetService_opslevel_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "api_token")
	base, _ :=getString(args, "base_url")
	id, _ :=getString(args, "service_id")
	url := base + "/api/services/" + id
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer " + token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Failed to execute request: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	return ok(string(body))
}