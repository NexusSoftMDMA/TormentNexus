package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListWorkflows(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "api_base")
	if base == "" {
		return err("api_base is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", base+"/workflows", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var workflows interface{}
	if e := json.Unmarshal(body, &workflows); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Workflows: %v", workflows))
}

func HandleRunWorkflow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "api_base")
	if base == "" {
		return err("api_base is required")
}

	workflowID, _ :=getString(args, "workflow_id")
	if workflowID == "" {
		return err("workflow_id is required")
}

	payload := map[string]string{"workflow_id": workflowID}
	jsonPayload, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal payload: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", base+"/workflows/run", bytes.NewReader(jsonPayload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(fmt.Sprintf("Workflow run response: %s", string(body)))
}