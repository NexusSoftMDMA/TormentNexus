//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		msg = "pong"
	}
	return success(msg)
}

func HandleFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url argument is required")
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	return success(fmt.Sprintf("fetched %s with status %d", url, resp.StatusCode))
}
