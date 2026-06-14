package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "http://localhost:8888"
	}
	u := fmt.Sprintf("%s/search?q=%s&format=json", baseURL, url.QueryEscape(query))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}