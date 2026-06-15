//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleDeepResearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	u := fmt.Sprintf("https://api.example.com/deep-research?q=%s", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("json decode failed: " + e.Error())
}

	summary, found := result["summary"].(string)
	if !found {
		summary = "No summary found"
	}
	return ok(summary)
}

func HandleResearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	if topic == "" {
		return err("topic is required")
}

	u := fmt.Sprintf("https://api.example.com/research?topic=%s", url.QueryEscape(topic))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&data)
	if e != nil {
		return err("json decode failed: " + e.Error())
}

	sources, found := data["sources"].([]interface{})
	if !found {
		sources = []interface{}{}
	}
	out, e := json.Marshal(sources)
	if e != nil {
		return err("marshal failed: " + e.Error())
}

	return ok(string(out))
}
