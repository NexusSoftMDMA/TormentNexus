//go:build ignore
// +build ignore

package tools

import "context"

func HandleReadFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	return ok("read file " + path)
}

func HandleListDirectory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	return ok("list directory " + path)
}
