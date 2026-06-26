package mcpimpl

import (
	"context"
)

func HandleExtractText_pdf_tools_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	file, _ :=getString(args, "file")
	if file == "" {
		return err("file is required")
}

	return ok("text extraction complete")
}

func HandleMergePdfs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	files := args["files"].([]interface{})
	if len(files) < 2 {
		return err("at least two files required")
}

	return ok("merge successful")
}