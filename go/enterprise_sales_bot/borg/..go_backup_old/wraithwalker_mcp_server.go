package tools

import (
	"context"
	"os"
)

func HandleListFixtureDirs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	root, _ :=getString(args, "root")
	if root == "" {
		root = "./fixtures"
	}
	entries, e := os.ReadDir(root)
	if e != nil {
		return err("failed to read root directory: " + e.Error())
	}
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())

	}
	return ok(map[string]interface{}{"directories": dirs})
}

}

func HandleGetFixtureFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path argument is required")
	}
	entries, e := os.ReadDir(path)
	if e != nil {
		return err("failed to read directory: " + e.Error())
	}
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())

	}
	return ok(map[string]interface{}{"files": files})
}
}