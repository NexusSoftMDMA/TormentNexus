package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleListMergeRequests(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "project_id")
	if projectID == "" {
		return err("project_id is required")
}

	token, _ :=getString(args, "access_token")
	if token == "" {
		return err("access_token is required")
}

	url := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/merge_requests", projectID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Private-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("non-200 response: " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	return success(string(body))
}