package tools

import (
	"context"
	"os"
	"strings"
)

func HandleReadFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		return err(readErr.Error())
}

	return ok(string(content))
}

func HandleWriteFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	content, _ :=getString(args, "content")
	writeErr := os.WriteFile(path, []byte(content), 0644)
	if writeErr != nil {
		return err(writeErr.Error())
}

	return ok("Successfully wrote content to " + path)
}

func HandleListDirectory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dirPath, _ :=getString(args, "path")
	entries, listErr := os.ReadDir(dirPath)
	if listErr != nil {
		return err(listErr.Error())
}

	var entryLines []string
	for _, entry := range entries {
		if entry.IsDir() {
			entryLines = append(entryLines, "[DIR] "+entry.Name())
		} else {
			entryLines = append(entryLines, "[FILE] "+entry.Name())
		}
	}
	return ok(strings.Join(entryLines, "\n"))
}

func HandleDeleteFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetPath, _ :=getString(args, "path")
	deleteErr := os.Remove(targetPath)
	if deleteErr != nil {
		return err(deleteErr.Error())
}

	return ok("Successfully deleted " + targetPath)
}