package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleFetchLink(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing 'url' argument")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	return ok(string(body))
}

func HandleCheckLink(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing 'url' argument")
}

	resp, e := http.DefaultClient.Head(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	resp.Body.Close()
	return success("status: " + resp.Status)
}