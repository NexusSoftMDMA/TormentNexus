//go:build ignore
// +build ignore

package tools

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func HandleWebclawScrape(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed: " + e.Error())
}

	defer resp.Body.Close()
	body
-reasoner (deepseek)*
}