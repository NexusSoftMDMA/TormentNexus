package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	status, _ :=getString(args, "status")
	url := fmt.Sprintf("https://api.atlas.com/v1/projects?limit=%d&status=%s", limit, status)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to list projects: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("Projects: %v", result))
}

func HandleCreateProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("project name is required")
}

	desc, _ :=getString(args, "description")
	body, _ := json.Marshal(map[string]string{"name": name, "description": desc})
	resp, e := http.DefaultClient.Post("https://api.atlas.com/v1/projects", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to create project: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return err("unexpected status: " + resp.Status)
}

	return success("Project created successfully")
}