package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "api_key")
	if token == "" {
		return err("api_key is required")
}

	companyID, _ :=getString(args, "company_id")
	if companyID == "" {
		return err("company_id is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.procore.com/v1.0/projects?company_id=%s", companyID), nil)
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
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}