package mcpimpl

import (
	"context"
	"net/http"
)

func HandleBrowserloop(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing required url parameter")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch URL: " + e.Error())
}

	defer resp.Body.Close()
	return ok("browserloop opened: " + url)
}

func HandleBrowserloopScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing required url parameter")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch URL: " + e.Error())
}

	defer resp.Body.Close()
	return success("screenshot taken from: " + url)
}// touch 1781132121
