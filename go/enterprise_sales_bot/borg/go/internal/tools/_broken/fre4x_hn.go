package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetTopStories_fre4x_hn(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit < 1 {
		limit = 10
	}
	resp, e := http.DefaultClient.Get("https://hacker-news.firebaseio.com/v0/topstories.json")
	if e != nil {
		return err("failed to fetch top stories")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var ids []int
	if e := json.Unmarshal(body, &ids); e != nil {
		return err("failed to parse ids")
}

	if limit > len(ids) {
		limit = len(ids)

	ids = ids[:limit]
	list, _ := json.Marshal(ids)
	return ok(string(list))
}

}

func HandleGetItem_fre4x_hn(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "id")
	url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch item")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var item map[string]interface{}
	if e := json.Unmarshal(body, &item); e != nil {
		return err("failed to parse item")
}

	if item == nil {
		return err("item not found")
}

	title, found := item["title"].(string)
	if !found {
		title = "no title"
	}
	return ok(fmt.Sprintf("ID: %d, Title: %s", id, title))
}