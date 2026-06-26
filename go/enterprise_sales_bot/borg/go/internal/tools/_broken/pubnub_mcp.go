package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandlePublish_pubnub_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	channel, _ :=getString(args, "channel"), getString(args, "message")
	pub, _ :=getString(args, "publish_key"), getString(args, "subscribe_key")
	apiURL := fmt.Sprintf("https://ps.pndsn.com/publish/%s/%s/0/%s/0/%s",
		pub, sub, channel, url.PathEscape(message))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("publish request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return ok(fmt.Sprintf("published to %s: %v", channel, result))
}

func HandleHistory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	channel, _ :=getString(args, "channel"), getString(args, "subscribe_key")
	count, _ :=getInt(args, "count")
	if count == 0 {
		count = 100
	}
	apiURL := fmt.Sprintf("https://ps.pndsn.com/v2/history/sub-key/%s/channel/%s?count=%d",
		sub, channel, count)
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("history request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return ok(fmt.Sprintf("history for %s: %v", channel, result))
}