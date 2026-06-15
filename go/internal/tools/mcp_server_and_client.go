//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		msg = "pong"
	}
	return ok(msg)
}

func HandleFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url argument is required")
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
	}
	return success("fetch completed")
}
