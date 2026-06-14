package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleListWorkflowRuns(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	tenantId, _ :=getString(args, "tenant_id")
	if apiKey == "" || tenantId == "" {
		return err("api_key and tenant_id required")
}

	url := fmt.Sprintf("https://app.hatchet.run/api/v1/tenants/%s/workflow-runs", tenantId)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	if resp.StatusCode != 200 {
		return err("API error: " + string(body))
}

	return success(string(body))
}

func HandleTriggerWorkflow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	workflowId, _ :=getString(args, "workflow_id")
	payload, found := args["payload"].(map[string]interface{})
	if !found {
		payload = map[string]interface{}{}
	}
	
	if apiKey == "" || workflowId == "" {
		return err("api_key and workflow_id required")
}

	bodyMap := map[string]interface{}{
		"workflowId": workflowId,
		"input":      payload,
	}
	bodyBytes, e := json.Marshal(bodyMap)
	if e != nil {
		return err("failed to encode payload")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://app.hatchet.run/api/v1/events", strings.NewReader(string(bodyBytes)))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		return err("API error: " + string(body))
}

	return success(string(body))
}

func HandleCancelWorkflowRun(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	runId, _ :=getString(args, "run_id")
	if apiKey == "" || runId == "" {
		return err("api_key and run_id required")
}

	url := fmt.Sprintf("https://app.hatchet.run/api/v1/workflow-runs/%s/cancel", runId)
	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		return err("API error: " + string(body))
}

	return success(string(body))
}// touch 1781132127
