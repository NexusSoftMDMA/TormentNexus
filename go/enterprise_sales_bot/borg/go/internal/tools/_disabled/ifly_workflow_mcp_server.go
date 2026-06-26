package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListWorkflows(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := os.Getenv("IFLY_WORKFLOW_API_BASE")
	if base == "" {
		base = "http://localhost:8080"
	}
	limit, _ :=getString(args, "limit")
	status, _ :=getString(args, "status")
	url := base + "/workflows?limit=" + limit + "&status=" + status
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch workflows: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
}

	var workflows []map[string]interface{}
	if e := json.Unmarshal(body, &workflows); e != nil {
		return err("failed to parse response: " + e.Error())
}

	data, _ := json.Marshal(workflows)
	return ok("found " + fmt.Sprint(len(workflows)) + " workflows: " + string(data))
}

func HandleGetWorkflowStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := os.Getenv("IFLY_WORKFLOW_API_BASE")
	if base == "" {
		base = "http://localhost:8080"
	}
	wfID, _ :=getString(args, "workflow_id")
	if wfID == "" {
		return err("workflow_id is required")
}

	url := base + "/workflows/" + wfID + "/status"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get workflow status: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
}

	var statusData map[string]interface{}
	if e := json.Unmarshal(body, &statusData); e != nil {
		return err("failed to parse response: " + e.Error())
}

	data, _ := json.Marshal(statusData)
	return ok("workflow status: " + string(data))
}