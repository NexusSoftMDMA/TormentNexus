package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleCreateDeployTicket(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	jiraURL, _ :=getString(args, "jira_url")
	token, _ :=getString(args, "token")
	project, _ :=getString(args, "project")
	summary, _ :=getString(args, "summary")
	description, _ :=getString(args, "description")

	issue := map[string]interface{}{
		"fields": map[string]interface{}{
			"project":     map[string]string{"key": project},
			"summary":     summary,
			"description": description,
			"issuetype":   map[string]string{"name": "Deploy"},
		},
	}
	body, e := json.Marshal(issue)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", jiraURL+"/rest/api/2/issue", strings.NewReader(string(body)))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return err("Jira returned status " + resp.Status)
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	key, found := result["key"].(string)
	if !found {
		return err("response missing key")
}

	return ok("Deploy ticket created: " + key)
}