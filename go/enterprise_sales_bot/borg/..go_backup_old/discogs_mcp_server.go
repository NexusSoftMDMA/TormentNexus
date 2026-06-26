package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearchDiscogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter required")
}

	u := "https://api.discogs.com/database/search?q=" + url.QueryEscape(query)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()

	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("json decode failed: " + e.Error())
}

	results, found := data["results"].([]interface{})
	if !found {
		return err("no results found")
}

	resultStr := fmt.Sprintf("Found %d results", len(results))
	return ok(resultStr)
}

func HandleGetRelease(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "release_id")
	if id == 0 {
		return err("release_id integer required")
}

	u := fmt.Sprintf("https://api.discogs.com/releases/%d", id)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()

	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("json decode failed: " + e.Error())
}

	title, _ := data["title"].(string)
	return ok("Release: " + title)
}