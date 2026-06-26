package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type wikiSearchResult struct {
	Count int `json:"count"`
	Value []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"value"`
}

func HandleAzureWikiSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	org := os.Getenv("AZURE_DEVOPS_ORG")
	proj := os.Getenv("AZURE_DEVOPS_PROJECT")
	token := os.Getenv("AZURE_DEVOPS_PAT")
	query, _ :=getString(args, "query")
	top, _ :=getInt(args, "top")
	if top <= 0 || top > 50 {
		top = 10
	}
	if query == "" {
		return err("query is required")
	}
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/wiki/wikis?searchText=%s&$top=%d&api-version=7.0", org, proj, query, top)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
	}
	req.SetBasicAuth("", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
	}
	var result wikiSearchResult
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("failed to decode response: %v", e))
	}
	output := fmt.Sprintf("Found %d wiki page(s):\n", result.Count)
	for _, v := range result.Value {
		output += fmt.Sprintf("- %s (%s)\n", v.Name, v.URL)

	return ok(output)
}
}