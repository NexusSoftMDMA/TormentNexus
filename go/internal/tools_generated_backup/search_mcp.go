package tools

import (
	"context"
	"encoding/json"
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

	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	apiURL := fmt.Sprintf("https://api.search.example.com?q=%s&limit=%d", url.QueryEscape(query), limit)
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Found results: %v", result))
}

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("pong")
}