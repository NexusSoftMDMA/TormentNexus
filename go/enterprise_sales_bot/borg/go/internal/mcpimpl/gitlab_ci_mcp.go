package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetPipelineStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "projectId")
	pipelineID, _ :=getInt(args, "pipelineId")
	if projectID == "" || pipelineID == 0 {
		return err("projectId and pipelineId are required")
}

	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		return err("GITLAB_TOKEN environment variable not set")
}

	url := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/pipelines/%d", projectID, pipelineID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("PRIVATE-TOKEN", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var pipeline struct {
		Status string `json:"status"`
	}
	if e := json.Unmarshal(body, &pipeline); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok("Pipeline status: " + pipeline.Status)
}