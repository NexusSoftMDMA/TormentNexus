package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandlePublish(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pubKey, _ :=getString(args, "publish_key")
	subKey, _ :=getString(args, "subscribe_key")
	channel, _ :=getString(args, "channel")
	message, _ :=getString(args, "message")

	msgEncoded := url.QueryEscape(message)
	urlStr := fmt.Sprintf("https://ps.pndsn.com/publish/%s/%s/0/%s/0/%s",
		pubKey, subKey, channel, msgEncoded)

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != 200 {
		return err("pubnub publish returned status " + resp.Status)
}

	return ok("Message published successfully")
}

func HandleChannels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	subKey, _ :=getString(args, "subscribe_key")
	urlStr := fmt.Sprintf("https://ps.pndsn.com/v2/presence/sub_key/%s/channel?uuid=mcp", subKey)

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	var result struct {
		Channels []string `json:"channels"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Channels: %v", result.Channels))
}