package mcpimpl

import (
	"context"
	"fmt"
	"os"
)

func HandleReadPdf(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "file_path")
	if filePath == "" {
		return err("file_path is required")
}

	data, e := os.ReadFile(filePath)
	if e != nil {
		return err(fmt.Sprintf("failed to read file: %v", e))
}

	return ok(string(data))
}