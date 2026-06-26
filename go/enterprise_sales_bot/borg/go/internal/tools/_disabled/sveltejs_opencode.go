package tools

import (
	"context"
	"fmt"
)

func HandleReadInstructions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "default"
	}
	msg := fmt.Sprintf("Instructions for '%s': Follow the Svelte OpenCode guidelines.", name)
	return ok(msg)
}

func HandleEditFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	content, _ :=getString(args, "content")
	if path == "" || content == "" {
		return err("path and content are required")
}

	msg := fmt.Sprintf("Edited file '%s' with %d bytes.", path, len(content))
	return success(msg)
}