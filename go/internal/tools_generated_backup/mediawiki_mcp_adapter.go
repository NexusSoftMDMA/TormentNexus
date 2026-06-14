package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchWiki(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	apiURL := fmt.Sprintf("https://en.wikipedia.org/w/api.php?action=query&list=search&srsearch=%s&format=json", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err(fmt.Sprintf("API request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response failed: %v", e))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	results, found := data["query"].(map[string]interface{})["search"].([]interface{})
	if !found {
		return ok("[]")
}

	return ok(fmt.Sprintf("%v", results))
}

func HandleGetPage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	if title == "" {
		return err("title parameter is required")
}

	apiURL := fmt.Sprintf("https://en.wikipedia.org/w/api.php?action=query&prop=extracts&exintro&explaintext&titles=%s&format=json", url.QueryEscape(title))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err(fmt.Sprintf("API request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response failed: %v", e))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	pages, found := data["query"].(map[string]interface{})["pages"].(map[string]interface{})
	if !found {
		return err("page not found")
}

	for _, page := range pages {
		pageMap, found := page.(map[string]interface{})
		if found {
			extract, found := pageMap["extract"]
			if found {
				return ok(fmt.Sprintf("%v", extract))
}

			return ok("no extract available")

	}
	return err("page not found")
}
}