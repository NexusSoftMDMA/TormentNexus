package mcpimpl

import (
	"context"
	"os"
	"strings"
)

func HandleIndexDir(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dir, _ :=getString(args, "directory")
	if dir == "" {
		return err("directory argument required")
}

	entries, e := os.ReadDir(dir)
	if e != nil {
		return err("failed to read directory: " + e.Error())
}

	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())

	return ok(strings.Join(names, "\n"))
}
}