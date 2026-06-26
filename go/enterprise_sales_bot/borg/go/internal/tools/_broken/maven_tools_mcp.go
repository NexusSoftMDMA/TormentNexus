package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleSearchMaven_maven_tools_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://search.maven.org/solrsearch/select?q=%s&rows=10&wt=json", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to search: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Response struct {
			Docs []struct {
				Id         string `json:"id"`
				LatestVersion string `json:"latestVersion"`
			} `json:"docs"`
		} `json:"response"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	if len(result.Response.Docs) == 0 {
		return ok("no results found")
}

	var lines []string
	for _, doc := range result.Response.Docs {
		lines = append(lines, fmt.Sprintf("%s (latest: %s)", doc.Id, doc.LatestVersion))

	return ok(strings.Join(lines, "\n"))
}
}