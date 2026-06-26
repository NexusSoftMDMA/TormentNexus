package tools

import (
	"context"
	"os/exec"
)

func HandleRunCli(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmdStr, _ :=getString(args, "command")
	if cmdStr == "" {
		return err("command is required")
}

	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return err("command failed: " + e.Error())
}

	return success(string(out))
}