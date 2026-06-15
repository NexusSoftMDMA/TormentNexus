//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleGetJiraIssue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	issueKey, _ :=getString(args, "issue_key")
	if issueKey == "" {
		return err("issue_key required")
}

	base, _ :=getString(args, "jira_base")
	user, _ :=getString(args, "user")
	token, _ :=getString(args, "token")
	if base == "" || user == "" || token == "" {
		return err("jira_base, user, token required")
}

	u := fmt.Sprintf("%s/rest/api/2/issue/%s", base, url.PathEscape(issueKey))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("create request: " + e.Error())
}

	req.SetBasicAuth(user, token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read: " + e.Error())
	if resp.StatusCode != 200 {
		return err("API error: " + string(body))
}

	var data map[string]interface{	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse: " + e.Error())
	return ok(fmt.Sprintf("")
}


-reasoner (deepseek)*
}
}
}