package tools

import (
	"context"
	"os/exec"
)

func HandleOterm(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	out, e := exec.CommandContext(ctx, "sh", "-c", cmd).Output()
	if e != nil {
		return err("command failed: " + e.Error())
}

	return ok(string(out))
}