package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleGetURL_mvf1(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url parameter required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("http get failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}