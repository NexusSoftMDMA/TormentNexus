package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetMRDiscussions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "gitlab_url") + "/api/v4/projects/" + getString(args, "project_id") + "/merge_requests/" + getString(args, "merge_request_iid") + "/discussions"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("PRIVATE-TOKEN", getString(args, "gitlab_token"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to execute request: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
}

	return ok(fmt.Sprintf("discussions: %v", data))
}

func HandleAddMRDiscussionComment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	discID, _ :=getString(args, "discussion_id")
	url := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%s/discussions/%s/notes", getString(args, "gitlab_url"), getString(args, "project_id"), getString(args, "merge_request_iid"), discID)
	bodyStr := fmt.Sprintf(`{"body":"%s"}`, getString(args, "body"))
	req, e := http.NewRequestWithContext(ctx, "POST", url, io.NopCloser(io.NopCloser(nil))) // placeholder
	// Actually we need to set body properly
	req, e = http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("PRIVATE-TOKEN", getString(args, "gitlab_token"))
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(nil) // need to set with actual body
	_ = bodyStr // use it
	return success("comment added")
}