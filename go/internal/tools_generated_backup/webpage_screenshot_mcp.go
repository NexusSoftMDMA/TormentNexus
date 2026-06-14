package tools

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
)

func HandleScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	if urlStr == "" {
		return err("url is required")
}

	_, e := url.ParseRequestURI(urlStr)
	if e != nil {
		return err("invalid url")
}

	reqURL := "https://screenshot.1place.io/?url=" + url.QueryEscape(urlStr)
	resp, e := http.DefaultClient.Get(reqURL)
	if e != nil {
		return err("failed to fetch screenshot: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("screenshot service returned status: " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	encoded := base64.StdEncoding.EncodeToString(body)
	return ok(encoded)
}