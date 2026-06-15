package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	org, _ :=getString(args, "organization")
	if org == "" {
		return err("organization is required")
}

	u := fmt.Sprintf("https://dev.azure.com/%s/_apis/projects?api-version=6.0", url.PathEscape(org))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Value []struct {
			Name string `json:"name"`
		} `json:"value"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
}

	names := make([]string, len(result.Value))
	for i, p := range result.Value {
		names[i] = p.Name
	}
	return ok(fmt.Sprintf("Projects: %v", names))
}