package tools

import (
	"context"
	"os/exec"
)

func HandlePlaywrightCli(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	out, e := exec.CommandContext(ctx, "npx", "playwright", cmd).Output()
	if e != nil {
		return err("failed to execute: " + e.Error())
}

	return ok(string(out))
}