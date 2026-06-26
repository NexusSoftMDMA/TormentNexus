package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleGetInfo_srunx(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read: " + e.Error())
}

	return success(string(body))
}

func HandlePing_srunx(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}