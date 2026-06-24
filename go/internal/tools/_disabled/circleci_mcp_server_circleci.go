package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleGetPipeline(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectSlug, _ :=getString(args, "project_slug")
	if projectSlug == "" {
		return err("project_slug is required")
}

	url := fmt.Sprintf("https://circleci.com/api/v2/project/%s/pipeline", projectSlug)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	return success("Successfully retrieved pipeline data for " + projectSlug)
}

func HandleTriggerPipeline(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectSlug, _ :=getString(args, "project_slug")
	branch, _ :=getString(args, "branch")
	if projectSlug == "" || branch == "" {
		return err("project_slug and branch are required")
}

	url := fmt.Sprintf("https://circleci.com/api/v2/project/%s/pipeline", projectSlug)
	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return err(fmt.Sprintf("Failed to trigger pipeline: %d", resp.StatusCode))
}

	return ok("Pipeline triggered successfully for branch " + branch)
}// touch 1781132122
