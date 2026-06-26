package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

func HandleGoogleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	u := "https://www.googleapis.com/customsearch/v1?q=" + url.QueryEscape(query)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to perform search: " + e.Error())
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

	return ok("search results obtained")
}

func HandleGoogleTranslate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	target, _ :=getString(args, "target")
	if text == "" || target == "" {
		return err("text and target parameters are required")
}

	u := "https://translation.googleapis.com/language/translate/v2?q=" + url.QueryEscape(text) + "&target=" + url.QueryEscape(target)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to translate: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse translation: " + e.Error())
}

	return ok("translation completed")
}