package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleSearchRepos_czlonkowski(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url := fmt.Sprintf("https://api.github.com/search/repositories?q=topic:%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch: " + e.Error())
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
		return err("decode error: " + e.Error())
}

	if len(result.Items) == 0 {
		return ok("No repositories found.")
}

	output := ""
	for _, item := range result.Items {
		output += fmt.Sprintf("- [%s](%s): %s\n", item.FullName, item.HTMLURL, item.Description)

	return success(output)
}
}