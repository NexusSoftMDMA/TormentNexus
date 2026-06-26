package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleSearchBookmarks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	query, _ :=getString(args, "query")
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 20
	}
	baseURL := "https://api.raindrop.io/rest/v1"
	url := baseURL + "/raindrops/0?sort=-created&search=" + query + "&perpage=" + fmt.Sprint(limit)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API error: " + resp.Status)
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	items, found := result["items"].([]interface{})
	if !found {
		return ok("no items found")
}

	return ok(fmt.Sprintf("found %d bookmarks", len(items)))
}

func HandleGetCollections(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	baseURL := "https://api.raindrop.io/rest/v1"
	url := baseURL + "/collections"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API error: " + resp.Status)
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	items, found := result["items"].([]interface{})
	if !found {
		return ok("no collections found")
}

	return ok(fmt.Sprintf("found %d collections", len(items)))
}