package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListWorkflows(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("base_url is required")
}

	url := fmt.Sprintf("%s/api/v1/workflows", baseURL)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch workflows: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}

func HandleExecuteWorkflow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	workflowID, _ :=getString(args, "workflow_id")
	if baseURL == "" || workflowID == "" {
		return err("base_url and workflow_id are required")
}

	url := fmt.Sprintf("%s/api/v1/workflows/trigger?workflowId=%s", baseURL, workflowID)
	resp, e := http.DefaultClient.Post(url, "application/json", nil)
	if e != nil {
		return err("failed to trigger workflow: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}