package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
)

func HandleSearchSlack(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		return err("missing SLACK_TOKEN env")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://slack.com/api/search.messages?query="+query, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	if result["ok"] != true {
		return err("Slack API error")
}

	msg := result["messages"].(map[string]interface{})["matches"]
	return ok("found " + formatCount(msg))
}

func HandleSearchReddit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://www.reddit.com/search.json?q="+query, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("User-Agent", "UnClick-MCP/1.0")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	data, found := result["data"].(map[string]interface{})
	if !found {
		return err("unexpected Reddit response")
}

	children, found := data["children"].([]interface{})
	if !found {
		return err("unexpected Reddit response")
}

	return ok("found " + formatCount(len(children)) + " results")
}

func formatCount(v interface{}) string {
	if c, found := v.(float64); found {
		return fmt.Sprintf("%.0f", c)
}

	if c, found := v.(int); found {
		return fmt.Sprintf("%d", c)
}

	return "unknown"
}