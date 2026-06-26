package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleListRepos_helm_chart_cli_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repoURL, _ :=getString(args, "repo_url")
	if repoURL == "" {
		repoURL = "https://helm.sh"
	}
	resp, e := http.DefaultClient.Get(repoURL + "/index.yaml")
	if e != nil {
		return err("failed to fetch repo index: " + e.Error())
}

	defer resp.Body.Close()
	var index struct {
		Entries map[string]interface{} `json:"entries"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&index); e != nil {
		return err("failed to parse index: " + e.Error())
}

	names := make([]string, 0, len(index.Entries))
	for name := range index.Entries {
		names = append(names, name)

	return success(fmt.Sprintf("Found %d charts: %v", len(names), names))
}

}

func HandleSearchChart(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return ok("no query provided")
}

	searchURL := fmt.Sprintf("https://hub.helm.sh/api/search?q=%s", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(searchURL)
	if e != nil {
		return err("search request failed: " + e.Error())
}

	defer resp.Body.Close()
	var results struct {
		Results []map[string]interface{} `json:"results"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&results); e != nil {
		return err("failed to parse search results: " + e.Error())
}

	return success(fmt.Sprintf("Found %d results for %q", len(results.Results), query))
}