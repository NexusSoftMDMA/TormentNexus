package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleTopStories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 30
	}
	resp, e := http.DefaultClient.Get("https://hacker-news.firebaseio.com/v0/topstories.json")
	if e != nil {
		return err("failed to get stories: " + e.Error())
}

	defer resp.Body.Close()
	var ids []int
	if e := json.NewDecoder(resp.Body).Decode(&ids); e != nil {
		return err("failed to decode ids: " + e.Error())
}

	if len(ids) > limit {
		ids = ids[:limit]
	}
	items := make([]map[string]interface{}, 0, len(ids))
	for _, id := range ids {
		item, e := fetchItem(id)
		if e != nil {
			continue
		}
		items = append(items, item)

	data, e := json.Marshal(items)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	return ok(string(data))
}

}

func HandleGetItem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "id")
	if id <= 0 {
		return err("id is required")
}

	item, e := fetchItem(id)
	if e != nil {
		return err("fetch error: " + e.Error())
}

	data, e := json.Marshal(item)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	return ok(string(data))
}

func fetchItem(id int) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()
	var item map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&item); e != nil {
		return nil, e
	}
	return item, nil
}