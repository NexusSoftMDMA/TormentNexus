package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleHttpBinGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://httpbin.org/get")
	if e != nil {
		return err("failed to fetch")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body")
}

	return ok(string(body))
}

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}