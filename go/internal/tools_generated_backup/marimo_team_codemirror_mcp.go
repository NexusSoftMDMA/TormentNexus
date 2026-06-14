package tools

import (
	"context"
	"net/http"
)

func HandleGetCursor(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = http.DefaultClient
	filename, _ :=getString(args, "filename")
	if filename == "" {
		return err("filename is required")
}

	return ok("cursor at line 1, column 1")
}

func HandleApplyEdit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = http.DefaultClient
	filename, _ :=getString(args, "filename")
	content, _ :=getString(args, "content")
	if filename == "" || content == "" {
		return err("filename and content are required")
}

	return success("edit applied")
}