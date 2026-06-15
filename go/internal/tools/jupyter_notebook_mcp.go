//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleListNotebooks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "jupyter_url")
	token, _ :=getString(args, "token")
	if url == "" {
		return err("jupyter_url is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", url+"/api/contents?type=notebook", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	if token != "" {
		req.Header.Set("Authorization", "Token "+token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to")
}


-reasoner (deepseek)*
}
}
}