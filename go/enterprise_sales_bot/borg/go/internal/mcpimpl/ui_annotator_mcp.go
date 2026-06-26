package mcpimpl

import (
	"context"
)

func HandleAnnotate_ui_annotator_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	msg := "Annotated URL: " + url
	return ok(msg)
}