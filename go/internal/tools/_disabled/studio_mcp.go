package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.studiomcp.example.com/projects")
	if e != nil {
		return err("failed to fetch projects: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var projects []map[string]interface{}
	if e := json.Unmarshal(body, &projects); e != nil {
		return err("failed to parse projects: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d projects", len(projects)))
}

func HandleGetProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing project id")
}

	url := "https://api.studiomcp.example.com/projects/" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch project: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var project map[string]interface{}
	if e := json.Unmarshal(body, &project); e != nil {
		return err("failed to parse project: " + e.Error())
}

	return ok(fmt.Sprintf("Project: %v", project))
}