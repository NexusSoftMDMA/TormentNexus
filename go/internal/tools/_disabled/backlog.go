package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleBacklogListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	apiKey, _ :=getString(args, "api_key")
	if baseURL == "" || apiKey == "" {
		return err("base_url and api_key are required")
	}
	url := strings.TrimRight(baseURL, "/") + "/api/v2/projects?apiKey=" + apiKey
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API error: " + resp.Status)
	}
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error: " + e.Error())
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return success(string(data))
}

func HandleBacklogCreateIssue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	apiKey, _ :=getString(args, "api_key")
	projectID, _ :=getString(args, "project_id")
	summary, _ :=getString(args, "summary")
	if baseURL == "" || apiKey == "" || projectID == "" || summary == "" {
		return err("base_url, api_key, project_id, summary are required")
	}
	bodyMap := map[string]interface{}{
		"projectId": projectID,
		"summary":   summary,
	}
	body, _ := json.Marshal(bodyMap)
	url := strings.TrimRight(baseURL, "/") + "/api/v2/issues?apiKey=" + apiKey
	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewReader(body))
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return err("API error: " + resp.Status)
	}
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error: " + e.Error())
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return success(string(data))
}