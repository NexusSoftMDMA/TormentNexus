package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetIssue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	key, _ :=getString(args, "api_key")
	id, _ :=getString(args, "issue_id")
	if id == "" {
		return err("issue_id required")
}

	url := base + "/issues/" + id + ".json?key=" + key
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("request error: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("json error: " + e.Error())
}

	out, _ := json.Marshal(data)
	return success(string(out))
}

func HandleListIssues(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	key, _ :=getString(args, "api_key")
	project, _ :=getString(args, "project_id")
	status, _ :=getString(args, "status_id")
	url := base + "/issues.json?key=" + key
	if project != "" {
		url += "&project_id=" + project
	}
	if status != "" {
		url += "&status_id=" + status
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("request error: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("json error: " + e.Error())
}

	out, _ := json.Marshal(data)
	return success(string(out))
}