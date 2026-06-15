package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetTopStories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://hacker-news.firebaseio.com/v0/topstories.json")
	if e != nil {
		return err("failed to fetch top stories: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var ids []int
	if e := json.Unmarshal(body, &ids); e != nil {
		return err("failed to parse top stories: " + e.Error())
}

	return ok(fmt.Sprintf("top stories (%d total): %v", len(ids), ids))
}

func HandleGetItem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "id")
	url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id)

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch item: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var item map[string]interface{}
	if e := json.Unmarshal(body, &item); e != nil {
		return err("failed to parse item: " + e.Error())
}

	return ok(fmt.Sprintf("item %d: %v", id, item))
}