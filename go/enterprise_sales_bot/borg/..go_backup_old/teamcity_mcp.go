package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "serverUrl")
	authToken, _ :=getString(args, "authToken")
	if serverURL == "" || authToken == "" {
		return err("serverUrl and authToken required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", serverURL+"/app/rest/projects", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	projects, found := result["project"].([]interface{})
	if !found {
		return err("unexpected response format")
}

	names := make([]string, 0, len(projects))
	for _, p := range projects {
		if proj, found := p.(map[string]interface{}); found {
			if name, found := proj["name"].(string); found {
				names = append(names, name)

		}
	}
	return ok(fmt.Sprintf("Found %d projects: %v", len(names), names))
}
}