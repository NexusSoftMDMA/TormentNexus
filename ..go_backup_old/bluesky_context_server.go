package tools

import (
	"context"
	"net/http"
)

func HandleGetPosts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username required")
}

	url := "https://bsky.social/xrpc/app.bsky.feed.getAuthorFeed?actor=" + username
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API returned status " + resp.Status)
}

	return success("Fetched posts for " + username)
}