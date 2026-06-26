package mcpimpl

import (
	"context"
	"io"
	"net/http"
	"time"
)

func HandleGetTime_mcptools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok(time.Now().Format(time.RFC3339))
}

func HandleFetchURL_mcptools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
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
		return err("failed to read body: " + e.Error())
}

	return ok(string(body))
}