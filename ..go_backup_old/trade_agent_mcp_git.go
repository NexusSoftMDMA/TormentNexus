package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearchRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	apiURL := fmt.Sprintf("https://api.github.com/search/repositories?q=%s", url.QueryEscape(query))
	req, e := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("GitHub API returned status %d", resp.StatusCode))
}

	var result struct {
		Items []struct {
			FullName string `json:"full_name"`
			HTMLURL  string `json:"html_url"`
			Description string `json:"description"`
		} `json:"items"`
	}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	if len(result.Items) == 0 {
		return success("no repositories found")
}

	response := fmt.Sprintf("Found %d repositories:\n", len(result.Items))
	for _, item := range result.Items {
		response += fmt.Sprintf("- %s: %s\n", item.FullName, item.HTMLURL)

	return success(response)
}
}