package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleListContainers_portainer_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := os.Getenv("PORTAINER_URL")
	token := os.Getenv("PORTAINER_TOKEN")
	if base == "" || token == "" {
		return err("PORTAINER_URL and PORTAINER_TOKEN must be set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/endpoints/1/docker/containers/json?all=true", base), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var containers []interface{}
	if e = json.NewDecoder(resp.Body).Decode(&containers); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d containers", len(containers)))
}

func HandleListStacks_portainer_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := os.Getenv("PORTAINER_URL")
	token := os.Getenv("PORTAINER_TOKEN")
	if base == "" || token == "" {
		return err("PORTAINER_URL and PORTAINER_TOKEN must be set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/stacks", base), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var stacks []interface{}
	if e = json.NewDecoder(resp.Body).Decode(&stacks); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d stacks", len(stacks)))
}