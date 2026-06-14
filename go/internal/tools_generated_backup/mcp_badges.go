package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleGetBadge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
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
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}