package mcpimpl

import (
	"context"
	"os"
)

func HandleReadFile_codanna(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	data, e := os.ReadFile(path)
	if e != nil {
		return err("failed to read file: " + e.Error())
}

	return success(string(data))
}