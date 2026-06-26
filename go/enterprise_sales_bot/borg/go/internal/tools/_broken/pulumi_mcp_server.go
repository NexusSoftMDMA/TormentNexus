package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListStacks_pulumi_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	org, _ :=getString(args, "organization")
	project, _ :=getString(args, "project")
	if org == "" || project == "" {
		return err("organization and project are required")
}

	url := fmt.Sprintf("https://api.pulumi.com/api/stacks?organization=%s&project=%s", org, project)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "token "+getString(args, "token"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Stacks []map[string]interface{} `json:"stacks"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok("Found " + fmt.Sprint(len(result.Stacks)) + " stacks")
}

func HandleGetStack(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	org, _ :=getString(args, "organization")
	project, _ :=getString(args, "project")
	stack, _ :=getString(args, "stack")
	if org == "" || project == "" || stack == "" {
		return err("organization, project, and stack are required")
}

	url := fmt.Sprintf("https://api.pulumi.com/api/stacks/%s/%s/%s", org, project, stack)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "token "+getString(args, "token"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok("Stack details retrieved")
}