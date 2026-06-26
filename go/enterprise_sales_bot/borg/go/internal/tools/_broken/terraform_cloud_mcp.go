package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleListWorkspaces_terraform_cloud_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	org, _ :=getString(args, "organization")
	if token == "" || org == "" {
		return err("token and organization required")
}

	url := fmt.Sprintf("https://app.terraform.io/api/v2/organizations/%s/workspaces", org)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/vnd.api+json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result struct {
		Data []struct {
			Attributes struct {
				Name string `json:"name"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	var names []string
	for _, w := range result.Data {
		names = append(names, w.Attributes.Name)

	return success("Workspaces: " + fmt.Sprintf("%v", names))
}

}

func HandleListRuns_terraform_cloud_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	org, _ :=getString(args, "organization")
	workspaceID, _ :=getString(args, "workspace_id")
	if token == "" || org == "" || workspaceID == "" {
		return err("token, organization, workspace_id required")
}

	url := fmt.Sprintf("https://app.terraform.io/api/v2/workspaces/%s/runs", workspaceID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/vnd.api+json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result struct {
		Data []struct {
			ID   string `json:"id"`
			Attributes struct {
				Status string `json:"status"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	var runs []string
	for _, r := range result.Data {
		runs = append(runs, fmt.Sprintf("%s (%s)", r.ID, r.Attributes.Status))

	return success("Runs: " + fmt.Sprintf("%v", runs))
}
}