package mcpimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

func HandleX_gitlab_mr_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	if project == "" {
		return err("project is required")
}

	state, _ :=getString(args, "state")
	if state == "" {
		state = "opened"
	}
	base := "https://gitlab.com/api/v4/projects/" + url.PathEscape(project) + "/merge_requests"
	u, e := url.Parse(base)
	if e != nil {
		return err("invalid URL")
}

	q := u.Query()
	q.Set("state", state)
	u.RawQuery = q.Encode()

	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch merge requests")
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e == nil {
		pretty, _ := json.MarshalIndent(data, "", "  ")
		return ok(string(pretty))
}

	return ok(string(body))
}