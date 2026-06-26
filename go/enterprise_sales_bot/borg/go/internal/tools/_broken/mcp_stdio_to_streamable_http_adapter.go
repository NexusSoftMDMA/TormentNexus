package mcpimpl

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func HandleHttpRequest_mcp_stdio_to_streamable_http_adapter(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	method, _ :=getString(args, "method")
	if method == "" {
		method = "GET"
	}
	bodyStr, _ :=getString(args, "body")
	var reqBody io.Reader
	if bodyStr != "" {
		reqBody = strings.NewReader(bodyStr)

	req, e := http.NewRequestWithContext(ctx, method, url, reqBody)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	bodyBytes, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(bodyBytes))
}
}