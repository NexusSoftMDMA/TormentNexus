package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// HandleGetRequest performs a GET request to the URL provided in args["url"].
func HandleGetRequest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return success(string(body))
}

// HandlePostRequest performs a POST request with JSON body from args["body"] to args["url"].
func HandlePostRequest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	bodyStr, _ :=getString(args, "body")
	if url == "" || bodyStr == "" {
		return err("url and body are required")
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(bodyStr))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return success(string(body))
}