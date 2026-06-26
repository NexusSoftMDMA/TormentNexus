package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListExperiments_simonecoelhosfo_optimizely_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "project_id")
	if projectID == "" {
		return err("project_id is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.optimizely.com/v2/experiments?project_id=%s", projectID), nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPTIMIZELY_ACCESS_TOKEN"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("Experiments: %v", result))
}

func HandleGetExperiment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	experimentID, _ :=getString(args, "experiment_id")
	if experimentID == "" {
		return err("experiment_id is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.optimizely.com/v2/experiments/%s", experimentID), nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPTIMIZELY_ACCESS_TOKEN"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("Experiment: %v", result))
}