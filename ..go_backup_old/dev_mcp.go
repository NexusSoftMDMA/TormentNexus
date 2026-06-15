package tools

import (
	"context"
	"os"
)

func HandleListFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	entries, e := os.ReadDir(path)
	if e != nil {
		return err("failed to list directory: " + e.Error())
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())

	return ok(names)
}

}

func HandleReadFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	data, e := os.ReadFile(path)
	if e != nil {
		return err("failed to read file: " + e.Error())
	}
	return ok(string(data))
}