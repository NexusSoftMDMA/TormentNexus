package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleWebSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	reqURL := "https://api.duckduckgo.com/?q=" + url.QueryEscape(query) + "&format=json"
	resp, e := http.DefaultClient.Get(reqURL)
	if e != nil {
		return err("failed to search: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	abstractText, found := data["AbstractText"].(string)
	if !found {
		abstractText = "No result found"
	}
	results, _ := json.Marshal(data["Results"])
	return ok(fmt.Sprintf("Abstract: %s\nResults: %s", abstractText, string(results)))
}

func HandleFetchURL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	uri, _ :=getString(args, "url")
	if uri == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(uri)
	if e != nil {
		return err("failed to fetch URL: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	if len(body) > 5000 {
		body = body[:5000]
	}
	return ok("Content:\n" + string(body))
}