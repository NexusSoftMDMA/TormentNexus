package tools

import (
	"context"
	"os"
	"path/filepath"
)

func HandleListReleases(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dir, _ :=getString(args, "path")
	if dir == "" {
		dir = "."
	}
	entries, e := os.ReadDir(dir)
	if e != nil {
		return err("failed to read directory: " + e.Error())
}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())

	return ok(names)
}

}

func HandleGetRelease(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dir, _ :=getString(args, "path")
	name, _ :=getString(args, "name")
	if dir == "" {
		dir = "."
	}
	fullPath := filepath.Join(dir, name)
	data, e := os.ReadFile(fullPath)
	if e != nil {
		return err("failed to read release: " + e.Error())
}

	return success(string(data))
}