package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleTestHTTP(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url parameter is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success("response: " + string(body))
}