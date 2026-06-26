package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleWeblateListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	apiToken, _ :=getString(args, "api_token")
	if baseURL == "" || apiToken == "" {
		return err("base_url and api_token are required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/projects/", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Token "+apiToken)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("HTTP %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	projects, found := result["results"].([]interface{})
	if !found {
		return err("unexpected response format")
}

	var names []string
	for _, p := range projects {
		proj, _ := p.(map[string]interface{})
		if name, found := proj["name"].(string); found {
			names = append(names, name)

	}
	return ok(fmt.Sprintf("Projects: %v", names))
}
}