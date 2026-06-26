package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleLinearSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	u := fmt.Sprintf("https://api.linear.app/graphql?query=%s", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
}

	return success(result)
}

func HandleSlackGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	channel, _ :=getString(args, "channel")
	if channel == "" {
		return err("channel is required")
}

	u := fmt.Sprintf("https://slack.com/api/conversations.history?channel=%s", url.QueryEscape(channel))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
}

	return success(result)
}