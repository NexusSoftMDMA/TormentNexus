package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleChat_converse_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	url, _ :=getString(args, "url")
	if message == "" || url == "" {
		return err("message and url are required")
}

	body, _ := json.Marshal(map[string]string{"message": message})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	reply, _ := result["response"].(string)
	if reply == "" {
		reply = fmt.Sprintf("%v", result)

	return ok(reply)
}

}

func HandleConsensus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	rawURLs, found := args["urls"]
	if !found {
		return err("urls is required")
}

	urls, found := rawURLs.([]interface{})
	if !found {
		return err("urls must be a list")
}

	if message == "" {
		return err("message is required")
}

	var replies []string
	for _, u := range urls {
		urlStr, _ := u.(string)
		if urlStr == "" {
			continue
		}
		resp, e := HandleChat(ctx, map[string]interface{}{"message": message, "url": urlStr})
		if e != nil {
			replies = append(replies, fmt.Sprintf("error from %s: %v", urlStr, e))
			continue
		}
		replies = append(replies, resp.Content)

	agg, _ := json.Marshal(replies)
	return ok(string(agg))
}
}