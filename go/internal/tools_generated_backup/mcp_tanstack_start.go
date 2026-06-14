package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetProjectInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectId, _ :=getString(args, "projectId")
	url := fmt.Sprintf("https://api.tanstack.com/start/projects/%s", projectId)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("Failed to get project: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("Failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("Project info: %+v", data))
}

func HandleStartProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("Project name is required")
}

	return success("Project '" + name + "' started successfully")
}