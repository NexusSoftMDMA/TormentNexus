package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleListModules(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	token, _ :=getString(args, "token")
	projectID, _ :=getInt(args, "project_id")
	u, e := url.Parse(baseURL + fmt.Sprintf("/api/v3/projects/%d/modules", projectID))
	if e != nil {
		return err(fmt.Sprintf("invalid url: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse json: %v", e))
}

	return ok(fmt.Sprintf("modules: %v", result))
}

func HandleGetTestCases(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	token, _ :=getString(args, "token")
	projectID, _ :=getInt(args, "project_id")
	moduleID, _ :=getInt(args, "module_id")
	u, e := url.Parse(baseURL + fmt.Sprintf("/api/v3/projects/%d/test-cases?moduleId=%d", projectID, moduleID))
	if e != nil {
		return err(fmt.Sprintf("invalid url: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse json: %v", e))
}

	return ok(fmt.Sprintf("test cases: %v", result))
}