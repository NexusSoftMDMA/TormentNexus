package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func ListRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	if topic == "" {
		topic = "anthropic-mcp"
	}
	apiURL := fmt.Sprintf("https://api.github.com/search/repositories?q=topic:%s", url.QueryEscape(topic))
	req, e := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Items []struct {
			FullName string `json:"full_name"`
		} `json:"items"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return success(fmt.Sprintf("Found %d repositories", len(result.Items)))
}

func GetRepo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	if owner == "" || repo == "" {
		return err("owner and repo are required")
}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", url.PathEscape(owner), url.PathEscape(repo))
	req, e := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var repoData struct {
		FullName string `json:"full_name"`
		Stars    int    `json:"stargazers_count"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&repoData); e != nil {
		return err("decode failed: " + e.Error())
}

	return success(fmt.Sprintf("Repo %s has %d stars", repoData.FullName, repoData.Stars))
}