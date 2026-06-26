package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetConfluencePage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiBase, _ :=getString(args, "apiBase")
	token, _ :=getString(args, "token")
	pageID, _ :=getString(args, "pageID")
	if apiBase == "" || token == "" || pageID == "" {
		return err("missing required arguments: apiBase, token, pageID")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/wiki/rest/api/content/%s", apiBase, pageID), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	title, found := result["title"].(string)
	if !found {
		return err("response missing title")
}

	return ok(fmt.Sprintf("Page title: %s", title))
}

func HandleGetJiraIssue_atlassian_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiBase, _ :=getString(args, "apiBase")
	token, _ :=getString(args, "token")
	issueKey, _ :=getString(args, "issueKey")
	if apiBase == "" || token == "" || issueKey == "" {
		return err("missing required arguments: apiBase, token, issueKey")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/rest/api/2/issue/%s", apiBase, issueKey), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	summary, found := result["fields"].(map[string]interface{})
	if !found {
		return err("response missing fields")
}

	title, found := summary["summary"].(string)
	if !found {
		return err("response missing summary")
}

	return ok(fmt.Sprintf("Issue summary: %s", title))
}