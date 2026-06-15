package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleDuckDuckGoSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("failed to call DuckDuckGo API: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	abstract, found := result["AbstractText"].(string)
	if !found {
		abstract = "No abstract available"
	}
	answer, found := result["Answer"].(string)
	if !found || answer == "" {
		answer = "No answer"
	}
	msg := fmt.Sprintf("Query: %s\nAnswer: %s\nAbstract: %s", query, answer, abstract)
	return ok(msg)
}

func HandleDuckDuckGoInstant(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return HandleDuckDuckGoSearch(ctx, args)
}