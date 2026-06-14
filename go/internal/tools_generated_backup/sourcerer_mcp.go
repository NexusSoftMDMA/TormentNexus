package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	u := "https://api.github.com/search/repositories?q=" + url.QueryEscape(query)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode: " + e.Error())
}

	count, found := data["total_count"].(float64)
	if !found {
		return err("invalid response")
}

	return ok("found " + url.QueryEscape(query) + " - " + toString(count))
}

func HandleGetInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repo, _ :=getString(args, "repo")
	if repo == "" {
		return err("repo is required")
}

	u := "https://api.github.com/repos/" + url.QueryEscape(repo)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode: " + e.Error())
}

	name, _ := data["full_name"].(string)
	desc, _ := data["description"].(string)
	return ok(name + ": " + desc)
}