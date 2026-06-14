package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetOpenrpcDoc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url parameter required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
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
		return err("failed to read body: " + e.Error())
}

	var doc map[string]interface{}
	e = json.Unmarshal(body, &doc)
	if e != nil {
		return ok(string(body))
}

	return success(string(body))
}

func HandleListOpenrpcMethods(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url parameter required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
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
		return err("failed to read body: " + e.Error())
}

	var doc map[string]interface{}
	e = json.Unmarshal(body, &doc)
	if e != nil {
		return err("invalid JSON: " + e.Error())
}

	methods, found := doc["methods"].([]interface{})
	if !found {
		return err("no methods array found")
}

	result, e := json.Marshal(methods)
	if e != nil {
		return err("failed to marshal methods: " + e.Error())
}

	return success(string(result))
}// touch 1781132136
