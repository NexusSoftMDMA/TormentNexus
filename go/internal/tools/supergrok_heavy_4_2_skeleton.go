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
	if msg == "" {
		return err("message is required")
}

	return ok("echo: " + msg)
}

func HandleFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("fetch failed: " + e.Error())
}

	defer resp.Body.Close()
	var data interface{}
	e = json.NewDecoder(resp.Body).Decode(&data)
	if e != nil {
		return err("invalid JSON: " + e.Error())
}

	return success(fmt.Sprintf("fetched %d bytes", resp.ContentLength))
}
