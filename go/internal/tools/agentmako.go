//go:build ignore
// +build ignore

package tools

import (
	"context"
	"io"
	"net/http"
	"strconv"
)

func HandleFetchUrl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	u, _ :=getString(args, "url")
	if u == "" {
		return err("url is required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
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

	return ok("fetched " + strconv.Itoa(len(body)) + " bytes")
}

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}
