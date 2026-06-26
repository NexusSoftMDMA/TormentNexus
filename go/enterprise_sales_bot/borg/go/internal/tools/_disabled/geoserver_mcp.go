package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleListWorkspaces(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("base_url is required")
}

	url := strings.TrimRight(baseURL, "/") + "/rest/workspaces.json"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result struct {
		Workspaces struct {
			Workspace []map[string]interface{} `json:"workspace"`
		} `json:"workspaces"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	names := []string{}
	for _, ws := range result.Workspaces.Workspace {
		if name, found := ws["name"].(string); found {
			names = append(names, name)

	}
	summary := fmt.Sprintf("Found %d workspaces: %s", len(names), strings.Join(names, ", "))
	return ok(summary)
}

}

func HandleListLayers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("base_url is required")
}

	url := strings.TrimRight(baseURL, "/") + "/rest/layers.json"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result struct {
		Layers struct {
			Layer []map[string]interface{} `json:"layer"`
		} `json:"layers"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	names := []string{}
	for _, l := range result.Layers.Layer {
		if name, found := l["name"].(string); found {
			names = append(names, name)

	}
	summary := fmt.Sprintf("Found %d layers: %s", len(names), strings.Join(names, ", "))
	return ok(summary)
}
}