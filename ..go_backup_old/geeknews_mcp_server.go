package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetNews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}
	url := "https://hacker-news.firebaseio.com/v0/topstories.json"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch news")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var ids []int
	if e := json.Unmarshal(body, &ids); e != nil {
		return err("failed to parse response")
}

	if limit > len(ids) {
		limit = len(ids)

	ids = ids[:limit]
	text := fmt.Sprintf("Top %d story IDs: %v", limit, ids)
	return ok(text)
}
}