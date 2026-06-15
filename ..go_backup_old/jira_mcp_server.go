package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func HandleGetIssue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "issue_key")
	if key == "" {
		return err("missing issue_key")
}

	base := os.Getenv("JIRA_URL")
	req, e := http.NewRequestWithContext(ctx, "GET", base+"/rest/api/3/issue/"+url.PathEscape(key), nil)
	if e != nil {
		return err(e.Error())
}

	req.SetBasicAuth(os.Getenv("JIRA_USER"), os.Getenv("JIRA_TOKEN"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("HTTP %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}

func HandleSearchIssues(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	jql, _ :=getString(args, "jql")
	if jql == "" {
		return err("missing jql")
}

	base := os.Getenv("JIRA_URL")
	u := fmt.Sprintf("%s/rest/api/3/search?jql=%s", base, url.QueryEscape(jql))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err(e.Error())
}

	req.SetBasicAuth(os.Getenv("JIRA_USER"), os.Getenv("JIRA_TOKEN"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("HTTP %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return ok(string(body))
}