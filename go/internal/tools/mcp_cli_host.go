//go:build ignore
// +build ignore

package tools

import (
	"context"
	"os/exec"
)

func HandleExecuteCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command is required")
}

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err("command failed: " + e.Error())
}

	return ok(string(output))
}

func HandleListDirectory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		path = "."
	}
	cmd := exec.CommandContext(ctx, "ls", "-la", path)
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err("list failed: " + e.Error())
}

	return ok(string(output))
}
