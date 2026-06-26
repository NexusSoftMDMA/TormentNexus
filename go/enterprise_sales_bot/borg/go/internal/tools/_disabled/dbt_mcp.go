package tools

import (
	"context"
	"net/http"
	"io"
	"encoding/json"
)

func HandleListModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "project_id")
	if projectID == "" {
		return err("project_id is required")
}

	resp, e := http.DefaultClient.Get("https://api.dbt.com/v1/projects/" + projectID + "/models")
	if e != nil {
		return err("failed to fetch models: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleRunModel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	modelID, _ :=getString(args, "model_id")
	if modelID == "" {
		return err("model_id is required")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.dbt.com/v1/models/"+modelID+"/run", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to run model: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON response: " + e.Error())
}

	return success("Model run initiated: " + modelID)
}