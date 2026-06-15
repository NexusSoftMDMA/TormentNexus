//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func HandleListDirectory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	entries, e := os.ReadDir(path)
	if e != nil {
		return err(fmt.Sprintf("read dir: %v", e))
}

	var out string
	for _, e := range entries {
		out += e.Name() + "\n"
	}
	return ok(out)
}

func HandleReadFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	data, e := os.ReadFile(filepath.Clean(path))
	if e != nil {
		return err(fmt.Sprintf("read file: %v", e))
}

	return ok(string(data))
}
