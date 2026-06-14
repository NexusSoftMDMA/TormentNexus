package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type guardianResponse struct {
	Response struct {
		Status  string          `json:"status"`
		Content json.RawMessage `json:"content"`
		Results json.RawMessage `json:"results"`
	} `json:"response"`
}

func HandleSearchArticles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	apiKey, _ :=getString(args, "apiKey")
	if query == "" || apiKey == "" {
		return err("query and apiKey are required")
}

	url := fmt.Sprintf("https://content.guardianapis.com/search?q=%s&api-key=%s&show-fields=all", query, apiKey)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to make request: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var gr guardianResponse
	if e := json.Unmarshal(body, &gr); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if gr.Response.Status != "ok" {
		return err("guardian API error: " + gr.Response.Status)
}

	return success(string(gr.Response.Results))
}

func HandleGetArticle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	apiKey, _ :=getString(args, "apiKey")
	if id == "" || apiKey == "" {
		return err("id and apiKey are required")
}

	url := fmt.Sprintf("https://content.guardianapis.com/%s?api-key=%s&show-fields=all", id, apiKey)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to make request: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var gr guardianResponse
	if e := json.Unmarshal(body, &gr); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if gr.Response.Status != "ok" {
		return err("guardian API error: " + gr.Response.Status)
}

	return success(string(gr.Response.Content))
}