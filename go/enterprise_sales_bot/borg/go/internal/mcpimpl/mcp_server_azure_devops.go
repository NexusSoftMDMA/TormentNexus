package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleListProjects_mcp_server_azure_devops(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	org, _ :=getString(args, "organization")
	pat, _ :=getString(args, "pat")
	url := fmt.Sprintf("https://dev.azure.com/%s/_apis/projects?api-version=6.0", org)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.SetBasicAuth("", pat)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var result struct {
		Value []struct {
			Name string `json:"name"`
		} `json:"value"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed")
}

	return ok(fmt.Sprintf("Projects: %v", result.Value))
}

func HandleListRepositories_mcp_server_azure_devops(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	org, _ :=getString(args, "organization")
	proj, _ :=getString(args, "project")
	pat, _ :=getString(args, "pat")
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/git/repositories?api-version=6.0", org, proj)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.SetBasicAuth("", pat)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var result struct {
		Value []struct {
			Name string `json:"name"`
		} `json:"value"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed")
}

	return ok(fmt.Sprintf("Repositories: %v", result.Value))
}