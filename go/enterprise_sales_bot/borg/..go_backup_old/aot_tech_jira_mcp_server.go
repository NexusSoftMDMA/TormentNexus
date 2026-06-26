package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetIssue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "baseUrl")
	key, _ :=getString(args, "issueKey")
	token, _ :=getString(args, "authToken")
	if key == "" {
		return err("issueKey is required")
	}
	url := fmt.Sprintf("%s/rest/api/2/issue/%s", base, key)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http request: %v", e))
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
	}
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)))
	}
	return ok(string(body))
}

func HandleSearchIssues(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "baseUrl")
	jql, _ :=getString(args, "jql")
	token, _ :=getString(args, "authToken")
	maxResults, _ :=getInt(args, "maxResults")
	if maxResults <= 0 {
		maxResults = 50
	}
	url := fmt.Sprintf("%s/rest/api/2/search?jql=%s&maxResults=%d", base, jql, maxResults)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http request: %v", e))
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
	}
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)))
	}
	return ok(string(body))
}