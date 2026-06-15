//go:build ignore
// +build ignore

package tools

import (
	"context"
	"os"
	"strings"
)

func HandleReadFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	data, e := os.ReadFile(path)
	if e != nil {
		return err("failed to read file: " + e.Error())
}

	return success(string(data))
}

func HandleListDirectory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	entries, e := os.ReadDir(path)
	if e != nil {
		return err("failed to list directory: " + e.Error())
}

	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())

	return ok(strings.Join(names, "\n"))
}
}
