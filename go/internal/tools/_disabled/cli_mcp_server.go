package tools

import (
	"context"
	"os/exec"
	"strings"
)

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return ok(message)
}

func HandleExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command is required")
}

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	output, e := cmd.Output()
	if e != nil {
		return err("execution failed: " + e.Error())
}

	return ok(strings.TrimSpace(string(output)))
}