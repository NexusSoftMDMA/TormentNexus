package tools

import (
	"context"
	"net/http"
)

func HandleCheckWeb(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		url = "https://example.com"
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	return ok("Status: " + resp.Status)
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	return ok("Echo: " + text)
}