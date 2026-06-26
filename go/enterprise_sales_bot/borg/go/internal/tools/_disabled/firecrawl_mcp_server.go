package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleScrape(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	apiKey, _ :=getString(args, "apiKey")
	apiBase, _ :=getString(args, "apiUrl")
	if apiBase == "" {
		apiBase = "http://localhost:3002"
	}
	reqURL := fmt.Sprintf("%s/v1/scrape?url=%s", apiBase, url)
	req, e := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

}

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	apiKey, _ :=getString(args, "apiKey")
	apiBase, _ :=getString(args, "apiUrl")
	if apiBase == "" {
		apiBase = "http://localhost:3002"
	}
	reqURL := fmt.Sprintf("%s/v1/search?q=%s", apiBase, query)
	req, e := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}
}