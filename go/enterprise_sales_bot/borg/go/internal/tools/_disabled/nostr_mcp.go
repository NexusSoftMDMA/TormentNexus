package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleRelayInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "relay_url")
	if url == "" {
		url = "https://relay.damus.io/"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var info interface{}
	if e := json.Unmarshal(body, &info); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok(string(body))
}

func HandleGetPublicKey(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("npub1dummykey0000000000000000000000000000000000")
}