package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleGetIssue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	issueKey, _ :=getString(args, "issueKey")
	if issueKey == "" {
		return err("issueKey is required")
}

	base := os.Getenv("ATLASSIAN_URL")
	email := os.Getenv("ATLASSIAN_EMAIL")
	token := os.Getenv("ATLASSIAN_TOKEN")
	req, e := http.NewRequestWithContext(ctx, "GET", base+"/rest/api/3/issue/"+issueKey, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(email, token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("parse error: " + e.Error())
}

	fields, found := data["fields"].(map[string]interface{})
	if found {
		summary, _ := fields["summary"].(string)
		return ok("Issue " + issueKey + ": " + summary)
}

	return ok("Issue " + issueKey + " retrieved")
}

func HandleSearchIssues(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	jql, _ :=getString(args, "jql")
	if jql == "" {
		return err("jql is required")
}

	base := os.Getenv("ATLASSIAN_URL")
	email := os.Getenv("ATLASSIAN_EMAIL")
	token := os.Getenv("ATLASSIAN_TOKEN")
	url := fmt.Sprintf("%s/rest/api/3/search?jql=%s", base, jql)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(email, token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("parse error: " + e.Error())
}

	issues, found := data["issues"].([]interface{})
	if found && len(issues) > 0 {
		return ok(fmt.Sprintf("Found %d issues", len(issues)))
}

	return ok("No issues found")
}