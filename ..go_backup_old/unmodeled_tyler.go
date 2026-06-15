package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleSearchClaudeMCPRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	perPage, _ :=getInt(args, "per_page", 10)
	url := fmt.Sprintf("https://api.github.com/search/repositories?q=topic:claude-mcp&per_page=%d", perPage)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Items []struct {
			FullName    string `json:"full_name"`
			Description string `json:"description"`
			HTMLURL     string `json:"html_url"`
		} `json:"items"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("GitHub API returned status " + fmt.Sprint(resp.StatusCode))
}

	if len(result.Items) == 0 {
		return ok("No repositories found with topic claude-mcp")
}

	var output string
	for _, repo := range result.Items {
		output += fmt.Sprintf("- %s: %s (%s)\n", repo.FullName, repo.Description, repo.HTMLURL)

	return ok(output)
}

}

func getInt(args map[string]interface{}, key string, defaultVal int) int {
	if v, found := args[key]; found {
		if f, found := v.(float64); found {
			return int(f)

	}
	return defaultVal
}
}