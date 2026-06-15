//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("missing message argument")
	}
	return ok(fmt.Sprintf("Echo: %s", msg))
}

func HandleFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url argument")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	buf := make([]byte, 1024)
	n, e := resp.Body.Read(buf)
	if e != nil && e.Error() != "EOF" {
		return err(e.Error())
	}
	return success(fmt.Sprintf("Status: %s, Body: %s", resp.Status, strings.TrimSpace(string(buf[:n]))))
}
