package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://codebeamer.example.com/api/v3/projects")
	if e != nil {
		return err("failed to fetch projects: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleGetProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "projectId")
	if id == "" {
		return err("projectId is required")
}

	url := "https://codebeamer.example.com/api/v3/projects/" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get project: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}