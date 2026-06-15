//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleWriteText(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	resp, e := http.DefaultClient.Post("https://example.com/write", "application/json", nil)
	if e != nil {
		return err("failed to call external API: " + e.Error())
}

	defer resp.Body.Close()
	return ok(fmt.Sprintf("written: %s (status %d)", text, resp.StatusCode))
}

func HandleReadFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	return success("content of " + path + " is placeholder")
}
