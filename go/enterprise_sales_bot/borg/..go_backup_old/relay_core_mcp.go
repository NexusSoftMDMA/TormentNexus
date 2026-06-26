package tools

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

func HandleGetStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	getString(args, "dummy")
	return ok("RelayCore MCP server is running")
}

func HandleForwardRequest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	method, _ :=getString(args, "method")
	if method == "" {
		method = "GET"
	}
	bodyStr, _ :=getString(args, "body")
	var req *http.Request
	var e error
	if bodyStr != "" {
		req, e = http.NewRequestWithContext(ctx, method, url, bytes.NewBufferString(bodyStr))
	} else {
		req, e = http.NewRequestWithContext(ctx, method, url, nil)

	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}
}