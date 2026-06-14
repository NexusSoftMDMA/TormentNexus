package tools

import (
	"context"
	"os/exec"
)

func HandleRunCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	out, e := exec.Command("sh", "-c", cmd).Output()
	if e != nil {
		return err("execution failed: " + e.Error())
}

	return ok(string(out))
}