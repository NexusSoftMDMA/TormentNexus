package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

func HandleCheckIgnore(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	dir := filepath.Dir(path)
	ignoreFile := filepath.Join(dir, ".mcpignore")
	data, e := os.ReadFile(ignoreFile)
	if e != nil {
		if os.IsNotExist(e) {
			return ok("No .mcpignore found, path is not ignored")
}

		return err("failed to read .mcpignore: " + e.Error())
}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		matched, e := filepath.Match(line, filepath.Base(path))
		if e != nil {
			continue
		}
		if matched {
			return ok("Path is ignored by pattern: " + line)

	}
	return ok("Path is not ignored")
}
}