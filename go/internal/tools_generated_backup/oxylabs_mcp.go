package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleFetchURL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url parameter is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch URL: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response body: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("received status %d: %s", resp.StatusCode, string(body)))
}

	return ok(fmt.Sprintf("Fetched content (%d bytes): %s", len(body), string(body)))
}