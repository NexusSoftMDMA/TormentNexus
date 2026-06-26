package tools

import (
	"context"
	"os/exec"
)

func HandleRunCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmdStr, _ :=getString(args, "command")
	if cmdStr == "" {
		return err("command is required")
}

	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
	out, e := cmd.Output()
	if e != nil {
		return err("failed to run command: " + e.Error())
}

	return success(string(out))
}