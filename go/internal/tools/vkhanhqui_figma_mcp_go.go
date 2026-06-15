//go:build ignore
// +build ignore

package tools

import (
	"context"
)

func HandleGetFileInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileKey, _ :=getString(args, "file_key")
	if fileKey == "" {
		return err("file_key is required")
}

	return ok(`{"file_key":"` + fileKey + `","name":"Sample Figma File","last_modified":"2025-01-01"}`)
}

func HandleGetComponents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileKey, _ :=getString(args, "file_key")
	if fileKey == "" {
		return err("file_key is required")
}

	return ok(`{"file_key":"` + fileKey + `","components":[{"id":"1","name":"Button"},{"id":"2","name":"Card"}]}`)
}
