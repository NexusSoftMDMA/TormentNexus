package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListWorkflows(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL := os.Getenv("N8N_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:5678"
	}
	apiKey := os.Getenv("N8N_API_KEY")
	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/rest/workflows", nil)
	if e != nil {
		return err(e.Error())
}

	if apiKey != "" {
		req.Header.Set("X-N8N-API-KEY", apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return err(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}

}

func HandleExecuteWorkflow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workflowID, _ :=getString(args, "workflow_id")
	if workflowID == "" {
		return err("workflow_id is required")
}

	baseURL := os.Getenv("N8N_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:5678"
	}
	apiKey := os.Getenv("N8N_API_KEY")
	req, e := http.NewRequestWithContext(ctx, "POST", baseURL+"/rest/workflows/"+workflowID+"/execute", nil)
	if e != nil {
		return err(e.Error())
}

	if apiKey != "" {
		req.Header.Set("X-N8N-API-KEY", apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return err(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}
}