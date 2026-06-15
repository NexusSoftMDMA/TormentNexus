package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleGetIssue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	issueKey, _ :=getString(args, "issueKey")
	apiURL, _ :=getString(args, "apiUrl")
	username, _ :=getString(args, "username")
	apiToken, _ :=getString(args, "apiToken")
	if issueKey == "" || apiURL == "" || username == "" || apiToken == "" {
		return err("missing required arguments: issueKey, apiUrl, username, apiToken")
}

	url := strings.TrimRight(apiURL, "/") + "/rest/api/2/issue/" + issueKey
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + apiToken))
	req.Header.Set("Authorization", "Basic "+auth)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return err(fmt.Sprintf("Jira API error (status %d): %s", resp.StatusCode, string(body)))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(string(body))
}

func HandleSearchIssues(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	jql, _ :=getString(args, "jql")
	apiURL, _ :=getString(args, "apiUrl")
	username, _ :=getString(args, "username")
	apiToken, _ :=getString(args, "apiToken")
	maxResults, _ :=getInt(args, "maxResults")
	if jql == "" || apiURL == "" || username == "" || apiToken == "" {
		return err("missing required arguments: jql, apiUrl, username, apiToken")
}

	if maxResults <= 0 {
		maxResults = 50
	}
	bodyPayload := map[string]interface{}{
		"jql":        jql,
		"maxResults": maxResults,
	}
	payloadBytes, e := json.Marshal(bodyPayload)
	if e != nil {
		return err("failed to marshal request body: " + e.Error())
}

	url := strings.TrimRight(apiURL, "/") + "/rest/api/2/search"
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(payloadBytes)))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + apiToken))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return err(fmt.Sprintf("Jira API error (status %d): %s", resp.StatusCode, string(body)))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(string(body))
}