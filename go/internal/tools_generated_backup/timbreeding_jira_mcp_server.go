package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func GetProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := os.Getenv("JIRA_URL")
	token := os.Getenv("JIRA_TOKEN")
	if base == "" || token == "" {
		return err("JIRA_URL and JIRA_TOKEN environment variables required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", base+"/rest/api/3/project", nil)
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
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("Jira returned " + resp.Status)
}

	var projects []interface{}
	json.Unmarshal(body, &projects)
	return success(fmt.Sprintf("Found %d projects", len(projects)))
}

func GetIssue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := os.Getenv("JIRA_URL")
	token := os.Getenv("JIRA_TOKEN")
	if base == "" || token == "" {
		return err("JIRA_URL and JIRA_TOKEN environment variables required")
}

	key, _ :=getString(args, "issueKey")
	if key == "" {
		return err("issueKey is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", base+"/rest/api/3/issue/"+key, nil)
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
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("Jira returned " + resp.Status)
}

	return success(string(body))
}