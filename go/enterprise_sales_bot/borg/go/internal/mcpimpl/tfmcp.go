package mcpimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleListTools_tfmcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok(`{"tools":[{"name":"get_top_stories","description":"Fetch top stories from HackerNews"}]}`)
}

func HandleCallTool_tfmcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
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
		return err("failed to parse response")
}

	if len(ids) > 5 {
		ids = ids[:5]
	}
	data, _ := json.Marshal(ids)
	return success(string(data))
}