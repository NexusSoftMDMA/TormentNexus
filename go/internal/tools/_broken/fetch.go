package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func performRequest(ctx context.Context, method, urlStr string, headers map[string]interface{}, bodyStr string) (string, error) {
	if urlStr == "" {
		return "", fmt.Errorf("url is required")
	}

	_, parseErr := url.Parse(urlStr)
	if parseErr != nil {
		return "", fmt.Errorf("invalid url: %v", parseErr)
	}

	var bodyReader io.Reader
	if bodyStr != "" {
		bodyReader = strings.NewReader(bodyStr)
	}

	req, reqErr := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
	if reqErr != nil {
		return "", fmt.Errorf("failed to create request: %v", reqErr)
	}

	for k, v := range headers {
		if strVal, found := v.(string); found {
			req.Header.Set(k, strVal)
		}
	}

	resp, doErr := http.DefaultClient.Do(req)
	if doErr != nil {
		return "", fmt.Errorf("request failed: %v", doErr)
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", fmt.Errorf("failed to read response body: %v", readErr)
	}

	result := map[string]interface{}{
		"status":     resp.StatusCode,
		"statusText": resp.Status,
		"headers":    resp.Header,
		"body":       string(respBody),
	}

	jsonBytes, jsonErr := json.MarshalIndent(result, "", "  ")
	if jsonErr != nil {
		return "", fmt.Errorf("failed to marshal result: %v", jsonErr)
	}

	return string(jsonBytes), nil
}

func HandleFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ := getString(args, "url")
	method, _ := getString(args, "method")
	if method == "" {
		method = "GET"
	}

	var headers map[string]interface{}
	if h, found := args["headers"].(map[string]interface{}); found {
		headers = h
	} else {
		headers = make(map[string]interface{})
	}

	bodyStr, _ := getString(args, "body")

	result, reqErr := performRequest(ctx, method, urlStr, headers, bodyStr)
	if reqErr != nil {
		return err(reqErr.Error())
	}

	return ok(result)
}

func HandleGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ := getString(args, "url")
	var headers map[string]interface{}
	if h, found := args["headers"].(map[string]interface{}); found {
		headers = h
	} else {
		headers = make(map[string]interface{})
	}

	result, reqErr := performRequest(ctx, "GET", urlStr, headers, "")
	if reqErr != nil {
		return err(reqErr.Error())
	}

	return ok(result)
}

func HandlePost(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ := getString(args, "url")
	bodyStr, _ := getString(args, "body")
	var headers map[string]interface{}
	if h, found := args["headers"].(map[string]interface{}); found {
		headers = h
	} else {
		headers = make(map[string]interface{})
	}

	result, reqErr := performRequest(ctx, "POST", urlStr, headers, bodyStr)
	if reqErr != nil {
		return err(reqErr.Error())
	}

	return ok(result)
}