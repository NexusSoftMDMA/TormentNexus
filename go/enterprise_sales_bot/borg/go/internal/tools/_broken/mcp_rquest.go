package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleHttpRequest_mcp_rquest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	method, _ :=getString(args, "method")
	if method == "" {
		method = "GET"
	}
	var bodyReader io.Reader
	if body := getString(args, "body"); body != "" {
		bodyReader = strings.NewReader(body)

	req, e := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(data)))
}

	return success(string(data))
}
}