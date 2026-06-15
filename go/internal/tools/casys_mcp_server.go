//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}

func HandleFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("json decode failed: %v", e))
}

	data, found := result["data"]
	if !found {
		return err("response missing 'data' field")
}

	return ok(fmt.Sprintf("%v", data))
}
