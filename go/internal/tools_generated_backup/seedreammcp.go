package tools

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	u := "https://api.seedreammcp.com/search?q=" + url.QueryEscape(query)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return success(string(body))
}

func HandleGetSeed(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	u := "https://api.seedreammcp.com/seeds/" + url.PathEscape(id)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return success(string(body))
}