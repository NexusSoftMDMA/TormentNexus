package tools

import (
	"context"
	"os"
	"path/filepath"
)

func HandleListFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dir, _ :=getString(args, "directory")
	if dir == "" {
		return err("directory argument required")
}

	entries, e := os.ReadDir(dir)
	if e != nil {
		return err("cannot read directory: " + e.Error())
}

	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())

	return ok(filepath.Join(dir, " has files: ") + join(names))
}

}

func HandleReadFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path argument required")
}

	data, e := os.ReadFile(path)
	if e != nil {
		return err("cannot read file: " + e.Error())
}

	return ok(string(data))
}

func join(s []string) string {
	out := ""
	for i, v := range s {
		if i > 0 {
			out += ", "
		}
		out += v
	}
	return out
}