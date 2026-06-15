package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListPosts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	url := fmt.Sprintf("https://api.eelzap.com/posts?limit=%d", limit)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}