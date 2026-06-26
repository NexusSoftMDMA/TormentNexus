package mcpimpl

import (
	"context"
	"os/exec"
)

func HandleConvert_markitdown(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filepath, _ :=getString(args, "filepath")
	if filepath == "" {
		return err("filepath is required")
}

	cmd := exec.CommandContext(ctx, "markitdown", filepath)
	output, e := cmd.Output()
	if e != nil {
		return err("conversion failed: " + e.Error())
}

	return ok(string(output))
}