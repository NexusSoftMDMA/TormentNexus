package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	token, _ :=getString(args, "access_token")
	if baseURL == "" || token == "" {
		return err("missing base_url or access_token")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/projects.json", nil)
	if e != nil {
		return err("request creation failed")
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed")
}

	return success(string(body))
}

func HandleGetProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	token, _ :=getString(args, "access_token")
	projectID, _ :=getInt(args, "project_id")
	if baseURL == "" || token == "" || projectID == 0 {
		return err("missing required args")
}

	url := fmt.Sprintf("%s/projects/%d.json", baseURL, projectID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("request creation failed")
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed")
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("JSON parse failed")
}

	return success(string(body))
}