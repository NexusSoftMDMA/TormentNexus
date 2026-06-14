package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleRunTest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "project_id")
	branch, _ :=getString(args, "branch")
	if projectID == "" {
		return err("project_id is required")
}

	body := fmt.Sprintf(`{"project_id":"%s","branch":"%s"}`, projectID, branch)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.lastest.io/run", strings.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Test run created: %v", result["run_id"]))
}

func HandleGetDiff(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	runID, _ :=getString(args, "run_id")
	if runID == "" {
		return err("run_id is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.lastest.io/diff/"+runID, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var diff map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&diff); e != nil {
		return err("decode failed: " + e.Error())
}

	return success(fmt.Sprintf("Diff: %v", diff))
}