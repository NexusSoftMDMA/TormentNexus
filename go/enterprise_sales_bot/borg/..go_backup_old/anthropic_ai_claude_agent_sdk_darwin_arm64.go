package tools

import (
	"context"
	"os/exec"
)

func HandleRunAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command argument required")
}

	cmd := exec.CommandContext(ctx, "claude-agent", command)
	out, e := cmd.Output()
	if e != nil {
		return err("failed: " + e.Error())
}

	return ok(string(out))
}

func HandleGetVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "claude-agent", "--version")
	out, e := cmd.Output()
	if e != nil {
		return err("failed: " + e.Error())
}

	return ok(string(out))
}