//go:build ignore
// +build ignore

package tools

import (
	"context"
	"io"
	"net/http"
	"time"
)

func HandleFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
	}
	timeout, _ :=getInt(args, "timeout")
	if timeout <= 0 {
		timeout = 30
	}
	client := http.DefaultClient
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: "+e.Error())
	}
	resp, e := client.Do(req)
	if e != nil {
		return err("failed to fetch: "+e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: "+e.Error())
	}
	return ok(string(body))
}
