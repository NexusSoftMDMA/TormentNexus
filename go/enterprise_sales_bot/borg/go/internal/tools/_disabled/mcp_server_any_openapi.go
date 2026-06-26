package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func HandleOpenApiExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	endpoint, _ :=getString(args, "endpoint")
	method, _ :=getString(args, "method")
	bodyStr, _ :=getString(args, "body")

	if baseURL == "" || endpoint == "" || method == "" {
		return err("base_url, endpoint, and method are required")
}

	url := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(endpoint, "/")

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

	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(respBody, &result); e != nil {
		return success(string(respBody))
}

	return ok(result)
}
}