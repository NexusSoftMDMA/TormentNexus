package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleListPipelines_buildkite_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	org, _ :=getString(args, "org")
	token, _ :=getString(args, "token")
	if org == "" || token == "" {
		return err("missing org or token")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.buildkite.com/v2/organizations/"+org+"/pipelines", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
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
		return err("bad status: " + resp.Status)
}

	return ok(string(body))
}

func HandleGetBuild(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	org, _ :=getString(args, "org")
	token, _ :=getString(args, "token")
	pipeline, _ :=getString(args, "pipeline")
	buildNumber, _ :=getString(args, "build_number")
	if org == "" || token == "" || pipeline == "" || buildNumber == "" {
		return err("missing required parameters")
}

	url := "https://api.buildkite.com/v2/organizations/" + org + "/pipelines/" + pipeline + "/builds/" + buildNumber
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
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
		return err("bad status: " + resp.Status)
}

	return ok(string(body))
}