package tools

import (
	"context"
	"os/exec"
	"time"
)

func HandleExecuteShell(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command is required")
}

	timeout, _ :=getInt(args, "timeout")
	if timeout <= 0 {
		timeout = 30
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err("execution failed: " + e.Error())
}

	return ok(string(output))
}