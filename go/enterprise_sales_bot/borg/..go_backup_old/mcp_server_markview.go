package tools

import (
	"context"
	"os/exec"
)

func HandlePreviewMarkdown(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "file_path")
	if path == "" {
		return err("file_path is required")
}

	cmd := exec.CommandContext(ctx, "open", path)
	if e := cmd.Run(); e != nil {
		return err("failed to open: " + e.Error())
}

	return ok("preview opened")
}