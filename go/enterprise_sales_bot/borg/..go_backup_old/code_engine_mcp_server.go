package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleListJobRuns(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "project_id")
	region, _ :=getString(args, "region")
	apiKey, _ :=getString(args, "api_key")
	if projectID == "" || region == "" || apiKey == "" {
		return err("missing required parameters: project_id, region, api_key")
}

	url := fmt.Sprintf("https://api.%s.codeengine.cloud.ibm.com/v2/projects/%s/job_runs", region, projectID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to execute request: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(string(body))
}