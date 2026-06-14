package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetIssue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "baseUrl")
	issueKey, _ :=getString(args, "issueKey")
	if issueKey == "" {
		return err("issueKey is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/rest/api/3/issue/"+issueKey, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("Issue %s: %v", issueKey, result))
}

func HandleSearchIssues(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "baseUrl")
	jql, _ :=getString(args, "jql")
	if jql == "" {
		return err("jql is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/rest/api/3/search?jql="+jql, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	return success(string(body))
}