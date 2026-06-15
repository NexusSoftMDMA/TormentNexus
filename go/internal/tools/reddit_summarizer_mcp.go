//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleRedditSummary(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	subreddit, _ :=getString(args, "subreddit")
	if subreddit == "" {
		return err("subreddit is required")
}

	limit, _ :=getInt(args, "limit")
	if limit <= 0 {

	}
	url := fmt.Sprintf("https://www.reddit.com/r/%s/hot.json?limit=%d", subreddit, limit)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request:")
}


-reasoner (deepseek)*
}