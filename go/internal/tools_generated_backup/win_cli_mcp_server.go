package tools

import (
	"context"
	"fmt"
	"os/exec"
)

func HandleRunCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	out, e := exec.CommandContext(ctx, "cmd", "/c", cmd).Output()
	if e != nil {
		return err(fmt.Sprintf("command failed: %v", e))
}

	return ok(string(out))
}

func HandleListProcesses(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	out, e := exec.CommandContext(ctx, "cmd", "/c", "tasklist").Output()
	if e != nil {
		return err(fmt.Sprintf("failed to list processes: %v", e))
}

	return ok(string(out))
}