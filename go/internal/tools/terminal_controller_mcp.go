//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"os/exec"
)

func HandleExecuteCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	out, e := exec.CommandContext(ctx, "sh", "-c", cmd).CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("execution failed: %s", e))
}

	return ok(string(out))
}

func HandleListDirectory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	out, e := exec.CommandContext(ctx, "ls", "-la", path).CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("failed to list directory: %s", e))
}

	return ok(string(out))
}
