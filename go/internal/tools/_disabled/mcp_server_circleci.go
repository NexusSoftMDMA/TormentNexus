package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListPipelines(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	slug, _ :=getString(args, "project_slug")
	if slug == "" {
		return err("project_slug is required")
}

	branch, _ :=getString(args, "branch")
	url := fmt.Sprintf("https://circleci.com/api/v2/project/%s/pipeline", slug)
	if branch != "" {
		url += "?branch=" + branch
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var data interface{}
	json.Unmarshal(body, &data)
	out, _ := json.MarshalIndent(data, "", "  ")
	return ok(string(out))
}