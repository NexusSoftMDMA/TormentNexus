package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetSecurityInsights(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resourceType, _ :=getString(args, "resource_type")
	url := "https://api.rad.security/insights?type=" + resourceType
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return success(result)
}

func HandleAnalyzeVulnerability(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cveID, _ :=getString(args, "cve_id")
	if cveID == "" {
		return err("cve_id is required")
}

	url := "https://api.rad.security/vulnerabilities/" + cveID
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return success(result)
}