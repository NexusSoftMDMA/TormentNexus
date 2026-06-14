package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type githubRepo struct {
	Name        string
	FullName    string `json:"full_name"`
	Description string
}

type searchResponse struct {
	Items []githubRepo `json:"items"`
}

func HandleListRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	perPage, _ :=getInt(args, "per_page")
	if perPage <= 0 {
		perPage = 5
	}
	u := fmt.Sprintf("https://api.github.com/search/repositories?q=topic:mcp-server&per_page=%d&sort=stars&order=desc", perPage)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("HTTP error: " + e.Error())
}

	defer resp.Body.Close()
	var sr searchResponse
	if e = json.NewDecoder(resp.Body).Decode(&sr); e != nil {
		return err("JSON decode error: " + e.Error())
}

	if len(sr.Items) == 0 {
		return ok("No repositories found with topic mcp-server.")
}

	out := "GitHub repositories with topic 'mcp-server':\n"
	for _, r := range sr.Items {
		out += fmt.Sprintf("- %s (%s): %s\n", r.FullName, r.Name, r.Description)

	return ok(out)
}
}