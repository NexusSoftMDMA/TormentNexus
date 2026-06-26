package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleCheckHeaders(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Head(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	headers := map[string]string{
		"Strict-Transport-Security": resp.Header.Get("Strict-Transport-Security"),
		"X-Frame-Options":          resp.Header.Get("X-Frame-Options"),
		"X-Content-Type-Options":   resp.Header.Get("X-Content-Type-Options"),
	}
	return success(fmt.Sprintf("Security headers: %v", headers))
}

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("ping failed: %v", e))
}

	resp.Body.Close()
	return ok(fmt.Sprintf("Status: %s", resp.Status))
}