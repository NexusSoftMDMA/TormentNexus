package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleFetchWebsite(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	if resp.StatusCode != 200 {
		return err("HTTP " + resp.Status + ": " + string(body))
}

	return ok(string(body))
}

func HandleCompareContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	current, _ :=getString(args, "current")
	previous, _ :=getString(args, "previous")
	if current == "" && previous == "" {
		return err("no content provided")
}

	if current == previous {
		return ok("unchanged")
}

	return ok("changed")
}