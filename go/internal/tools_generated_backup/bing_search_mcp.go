package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleBingSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://api.bing.microsoft.com/v7.0/search?q=%s", query)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Ocp-Apim-Subscription-Key", os.Getenv("BING_API_KEY"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %d %s", resp.StatusCode, string(body)))
}

	var result struct {
		WebPages struct {
			Value []struct {
				Name string `json:"name"`
				URL  string `json:"url"`
				Snippet string `json:"snippet"`
			} `json:"value"`
		} `json:"webPages"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	var output string
	for _, item := range result.WebPages.Value {
		output += fmt.Sprintf("- %s\n  %s\n  %s\n", item.Name, item.URL, item.Snippet)

	if output == "" {
		return ok("No results found.")
}

	return ok(output)
}
}