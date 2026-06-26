package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleSecureaudit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	hasHTTPS := strings.HasPrefix(url, "https://")
	result := fmt.Sprintf("URL: %s\nStatus: %d\nHTTPS: %v\nContent-Length: %d", url, resp.StatusCode, hasHTTPS, len(body))
	return ok(result)
}