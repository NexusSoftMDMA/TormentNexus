package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL := "https://datalore.jetbrains.com/api"
	resp, e := http.DefaultClient.Get(baseURL + "/projects")
	if e != nil {
		return err("failed to fetch projects")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}

func HandleGetProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "projectId")
	if projectID == "" {
		return err("projectId required")
}

	baseURL := "https://datalore.jetbrains.com/api"
	url := fmt.Sprintf("%s/projects/%s", baseURL, projectID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch project")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}