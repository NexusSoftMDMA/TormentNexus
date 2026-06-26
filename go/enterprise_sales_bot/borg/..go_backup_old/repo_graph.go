package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type repoInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Stars       int    `json:"stargazers_count"`
	Forks       int    `json:"forks_count"`
	Language    string `json:"language"`
}

func HandleGetRepoGraph(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repoURL, _ :=getString(args, "url")
	if repoURL == "" {
		return err("url is required")
}

	u, e := url.Parse(repoURL)
	if e != nil {
		return err("invalid URL: " + e.Error())
}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return err("URL must point to a GitHub repository (owner/repo)")
}

	owner, repo := parts[0], parts[1]
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("GitHub API returned status " + resp.Status)
}

	var info repoInfo
	if e := json.NewDecoder(resp.Body).Decode(&info); e != nil {
		return err("failed to decode response: " + e.Error())
}

	msg := fmt.Sprintf("Repo: %s\nDescription: %s\nStars: %d\nForks: %d\nLanguage: %s",
		info.Name, info.Description, info.Stars, info.Forks, info.Language)
	return success(msg)
}