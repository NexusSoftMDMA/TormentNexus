package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		query = "anthropic-mcp"
	}
	url := fmt.Sprintf("https://api.github.com/search/repositories?q=topic:%s&per_page=5", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch repos: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result struct {
		Items []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			HTMLURL     string `json:"html_url"`
		} `json:"items"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	if len(result.Items) == 0 {
		return ok("No repositories found with topic: " + query)
}

	var msg string
	for _, item := range result.Items {
		msg += fmt.Sprintf("- %s: %s\n  %s\n", item.Name, item.Description, item.HTMLURL)

	return success("Repositories found:\n" + msg)
}
}