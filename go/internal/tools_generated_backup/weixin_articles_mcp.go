package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchArticles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	u := "https://api.weixin.qq.com/cgi-bin/wxopen/search?q=" + url.QueryEscape(query)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	data, e := json.MarshalIndent(result, "", "  ")
	if e != nil {
		return err("marshal failed: " + e.Error())
}

	return ok("search results: " + string(data))
}