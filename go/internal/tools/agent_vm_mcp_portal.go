//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func HandleListTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tools := []map[string]interface{}{
		{"name": "example_tool", "description": "An example tool"},
	}
	return success(map[string]interface{}{"tools": tools})
}

func HandleCallTool(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	payload, _ :=getString(args, "payload")
	var body io.Reader
	if payload != "" {
		body = strings.NewReader(payload)

	req, e := http.NewRequestWithContext(ctx, "POST", url, body)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	respBytes, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var result interface{}
	if e := json.Unmarshal(respBytes, &result); e != nil {
		return err(e.Error())
}

	return success(result)
}
}
