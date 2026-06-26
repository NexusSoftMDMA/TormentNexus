package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleCallAPI(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	method, _ :=getString(args, "method")
	url, _ :=getString(args, "url")
	bodyStr, _ :=getString(args, "body")
	if url == "" {
		return err("url is required")
}

	var body io.Reader
	if bodyStr != "" {
		body = bytes.NewBufferString(bodyStr)

	req, e := http.NewRequestWithContext(ctx, method, url, body)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(data, &result); e != nil {
		return ok(string(data))
}

	return ok(result)
}

}

func HandleListEndpoints(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	if base == "" {
		return err("base_url is required")
}

	resp, e := http.DefaultClient.Get(base + "/openapi.json")
	if e != nil {
		return err("failed to fetch OpenAPI spec: " + e.Error())
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read spec: " + e.Error())
}

	var spec map[string]interface{}
	if e := json.Unmarshal(data, &spec); e != nil {
		return ok(string(data))
}

	return ok(spec)
}