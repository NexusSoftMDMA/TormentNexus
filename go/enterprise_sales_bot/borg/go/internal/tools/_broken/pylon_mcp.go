package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandlePylonInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	if project == "" {
		return err("project is required")
}

	url := "https://api.pylon.com/projects/" + project
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch project: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return success(string(data))
}

func HandlePylonList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	url := "https://api.pylon.com/projects?limit=" + string(limit)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to list projects: " + e.Error())
}

	defer resp.Body.Close()
	var data []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode list: " + e.Error())
}

	return success(string(data))
}