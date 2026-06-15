package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleTriggerWorkflow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	workflowID, _ :=getString(args, "workflow_id")
	ref, _ :=getString(args, "ref")
	if owner == "" || repo == "" || workflowID == "" || ref == "" {
		return err("Missing required arguments: owner, repo, workflow_id, ref")
}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/workflows/%s/dispatches", owner, repo, workflowID)
	body := map[string]string{"ref": ref}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("Failed to marshal request body: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Failed to send request: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		return err(fmt.Sprintf("GitHub API returned status %d", resp.StatusCode))
}

	return success("Workflow dispatch triggered successfully")
}