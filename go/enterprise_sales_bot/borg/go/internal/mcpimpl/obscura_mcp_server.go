package mcpimpl

import (
	"context"
	"io"
	"net/http"
	"regexp"
)

func HandleFetchAndConvert(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	if urlStr == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(urlStr)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(string(body), "")
	return ok("Converted content: " + text)
}