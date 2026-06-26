package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchYoutube(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	apiKey, _ :=getString(args, "apiKey")
	if query == "" {
		return err("query is required")
}

	if apiKey == "" {
		return err("apiKey is required")
}

	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?part=snippet&q=%s&key=%s", query, apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to search YouTube: " + e.Error())
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(data, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	items, found := result["items"].([]interface{})
	if !found {
		return err("no items found")
}

	return success(fmt.Sprintf("Found %d videos", len(items)))
}