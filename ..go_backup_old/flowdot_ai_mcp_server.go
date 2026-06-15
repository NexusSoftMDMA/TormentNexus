package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCreateWorkflowRun(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workflowID, _ :=getString(args, "workflow_id")
	apiKey, _ :=getString(args, "api_key")
	payloadStr, _ :=getString(args, "payload")

	url := fmt.Sprintf("https://api.flowdot.ai/workflows/%s/run", workflowID)
	body := map[string]string{"payload": payloadStr}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal payload")
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	return ok(fmt.Sprintf("Workflow run created: %v", result["id"]))
}