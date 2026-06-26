package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListPosts_ghost_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	site, _ :=getString(args, "site")
	key, _ :=getString(args, "key")
	url := fmt.Sprintf("%s/ghost/api/v3/content/posts/?key=%s", site, key)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch posts")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("invalid response")
}

	return ok(fmt.Sprintf("Posts: %v", result["posts"]))
}

func HandleGetPost_ghost_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	site, _ :=getString(args, "site")
	key, _ :=getString(args, "key")
	id, _ :=getString(args, "id")
	url := fmt.Sprintf("%s/ghost/api/v3/content/posts/%s/?key=%s", site, id, key)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch post")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("invalid response")
}

	return ok(fmt.Sprintf("Post: %v", result["posts"]))
}