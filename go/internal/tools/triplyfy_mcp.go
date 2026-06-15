//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "world"
	}
	return ok("Hello, " + name + "!")
}

func HandleFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	return ok(fmt.Sprintf("HTTP %d", resp.StatusCode))
}
