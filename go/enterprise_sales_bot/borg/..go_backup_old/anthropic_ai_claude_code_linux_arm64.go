package tools

import (
	"context"
	"os/exec"
)

func HandleClaudeVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "claude-code", "--version")
	out, e := cmd.Output()
	if e != nil {
		return err(e.Error())
}

	return ok(string(out))
}

func HandleClaudeCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	cmd := exec.CommandContext(ctx, "claude-code", command)
	out, e := cmd.Output()
	if e != nil {
		return err(e.Error())
}

	return ok(string(out))
}