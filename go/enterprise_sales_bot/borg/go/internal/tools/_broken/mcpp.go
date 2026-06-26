package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchMcpRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		query = "topic:mcp-framework"
	}
	url := fmt.Sprintf("https://api.github.com/search/repositories?q=%s", query)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result struct {
		Items []struct {
			FullName string `json:"full_name"`
		} `json:"items"`
	}
	if e = json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if len(result.Items) == 0 {
		return ok("No repositories found.")
}

	names := make([]string, 0, len(result.Items))
	for _, item := range result.Items {
		names = append(names, item.FullName)

	msg := fmt.Sprintf("Found %d repositories: %v", len(names), names)
	return success(msg)
}

}

func HandleUppercase_mcpp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("missing 'text' argument")
}

	return success("Uppercased: " + text)
}