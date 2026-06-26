package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetJiraIssue_jira_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	issueKey, _ :=getString(args, "issueKey")
	if issueKey == "" {
		return err("issueKey is required")
}

	baseURL, _ :=getString(args, "baseUrl")
	if baseURL == "" {
		return err("baseUrl is required")
}

	username, _ :=getString(args, "username")
	password, _ :=getString(args, "password")
	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/rest/api/2/issue/"+issueKey, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(username, password)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Issue: %v", result))
}

func HandleSearchJiraIssues(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	jql, _ :=getString(args, "jql")
	if jql == "" {
		return err("jql is required")
}

	baseURL, _ :=getString(args, "baseUrl")
	if baseURL == "" {
		return err("baseUrl is required")
}

	username, _ :=getString(args, "username")
	password, _ :=getString(args, "password")
	maxResults, _ :=getInt(args, "maxResults")
	if maxResults == 0 {
		maxResults = 50
	}
	url := fmt.Sprintf("%s/rest/api/2/search?jql=%s&maxResults=%d", baseURL, jql, maxResults)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(username, password)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	issues, found := result["issues"].([]interface{})
	if !found {
		return err("no issues found")
}

	return ok(fmt.Sprintf("Found %d issues", len(issues)))
}