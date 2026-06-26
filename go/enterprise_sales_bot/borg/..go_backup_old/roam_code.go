package tools

import (
	"context"
	"fmt"
	"os"
)

func HandleListFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dir, _ :=getString(args, "directory")
	if dir == "" {
		return err("directory argument required")
}

	entries, e := os.ReadDir(dir)
	if e != nil {
		return err("failed to read directory: " + e.Error())
}

	names := []string{}
	for _, entry := range entries {
		names = append(names, entry.Name())

	return ok(fmt.Sprintf("Files: %v", names))
}

}

func HandleGetFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path argument required")
}

	content, e := os.ReadFile(path)
	if e != nil {
		return err("failed to read file: " + e.Error())
}

	return ok(string(content))
}