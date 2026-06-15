//go:build ignore
// +build ignore

package tools

import (
	"context"
	"net/http"
)

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}

func HandleHealth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://httpbin.org/get")
	if e != nil {
		return err("health check failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status")
}

	return success("healthy")
}
