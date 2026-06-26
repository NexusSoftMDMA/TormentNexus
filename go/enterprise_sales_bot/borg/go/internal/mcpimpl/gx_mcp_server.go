package mcpimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleFetchReadme(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	if owner == "" || repo == "" {
		return err("owner and repo are required")
}

	url := "https://raw.githubusercontent.com/" + owner + "/" + repo + "/main/README.md"
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
		return err("failed to read body: " + e.Error())
}

	return ok(string(body))
}

func HandleListCategories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://raw.githubusercontent.com/awesome-mcp-servers/awesome-mcp-servers/main/README.md"
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
		return err("failed to read body: " + e.Error())
}

	result := map[string]interface{}{
		"readme_preview": string(body)[:min(len(body), 2000)],
	}
	jsonBytes, e := json.Marshal(result)
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	return success(string(jsonBytes))
}// touch 1781132127
