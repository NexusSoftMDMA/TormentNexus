package mcpimpl

import (
	"context"
	"os"
)

func HandleReadFile_editor_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	data, e := os.ReadFile(path)
	if e != nil {
		return err("failed to read file: " + e.Error())
}

	return ok(string(data))
}

func HandleWriteFile_editor_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	content, _ :=getString(args, "content")
	e := os.WriteFile(path, []byte(content), 0644)
	if e != nil {
		return err("failed to write file: " + e.Error())
}

	return ok("file written")
}