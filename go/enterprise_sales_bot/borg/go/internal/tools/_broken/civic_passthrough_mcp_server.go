package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandlePassthrough(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	method, _ :=getString(args, "method")
	if method == "" {
		method = "GET"
	}
	body, _ :=getString(args, "body")
	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)

	req, e := http.NewRequestWithContext(ctx, method, url, reqBody)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	headers, found := args["headers"].(map[string]interface{})
	if found {
		for k, v := range headers {
			req.Header.Set(k, fmt.Sprint(v))

	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(fmt.Sprintf("Status: %d\n%s", resp.StatusCode, string(respBody)))
}
}
}