package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	apiKey, _ :=getString(args, "api_key")
	if url == "" || apiKey == "" {
		return err("url and api_key are required")
}

	reqURL := fmt.Sprintf("https://api.urlbox.io/v1/%s?url=%s", apiKey, url)
	resp, e := http.DefaultClient.Get(reqURL)
	if e != nil {
		return err("failed to fetch screenshot: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}