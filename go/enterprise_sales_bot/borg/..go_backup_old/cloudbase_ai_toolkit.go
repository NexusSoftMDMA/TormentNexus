package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.cloudbase.ai/v1/models", nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	var data []map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response")
}

	return ok(fmt.Sprintf("Models: %+v", data))
}

func HandleExecuteWorkflow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workflowID, _ :=getString(args, "workflow_id")
	if workflowID == "" {
		return err("workflow_id required")
}

	input, _ :=getString(args, "input")
	payload := map[string]string{"input": input}
	bodyBytes, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request")
}

	url := fmt.Sprintf("https://api.cloudbase.ai/v1/workflows/%s/execute", workflowID)
	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(respBody)))
}

	return ok(fmt.Sprintf("Workflow executed: %s", string(respBody)))
}