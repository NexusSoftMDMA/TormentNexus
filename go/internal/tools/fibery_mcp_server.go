//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleFiberyApi(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	method, _ :=getString(args, "method")
	path, _ :=getString(args, "path")
	bodyStr, _ :=getString(args, "body")
	if apiKey == "" || method == "" || path == "" {
		return err("apiKey, method, and path are required")
}

	url := fmt.Sprintf("https://api.fibery.io/api/v2%s", path)
	var body *strings.Reader
	if bodyStr != "" {

	} else {
		body = strings.NewReader("")

	req, e := http.NewRequestWithContext(ctx, strings.ToUpper(method), url, body)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	if bodyStr != "" {
		req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp()


-reasoner (deepseek)*
}
}
}