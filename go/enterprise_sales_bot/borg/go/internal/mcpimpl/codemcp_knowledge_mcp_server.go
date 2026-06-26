package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleGetKnowledge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	if topic == "" {
		return err("topic is required")
}

	resp, e := http.DefaultClient.Get("https://api.example.com/knowledge?topic=" + url.QueryEscape(topic))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("json decode failed: %v", e))
}

	data, found := result["data"]
	if !found {
		return err("no data found")
}

	return ok(fmt.Sprintf("Knowledge: %v", data))
}

func HandleSearchKnowledge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	resp, e := http.DefaultClient.Get("https://api.example.com/search?q=" + url.QueryEscape(query))
	if e != nil {
		return err(fmt.Sprintf("search failed: %v", e))
}

	defer resp.Body.Close()
	var results []interface{}
	if e := json.NewDecoder(resp.Body).Decode(&results); e != nil {
		return err(fmt.Sprintf("json decode failed: %v", e))
}

	if len(results) == 0 {
		return err("no results found")
}

	return success(fmt.Sprintf("Found %d results", len(results)))
}