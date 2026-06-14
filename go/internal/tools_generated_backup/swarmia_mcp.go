package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	orgID, _ :=getString(args, "organization_id")
	if orgID == "" {
		return err("organization_id is required")
}

	url := fmt.Sprintf("https://api.swarmia.com/v1/organizations/%s/projects", orgID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
}

	return success(fmt.Sprintf("Projects: %v", result))
}