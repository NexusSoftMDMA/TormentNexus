package mcpimpl

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func HandleDoubletickCli(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command is required")
}

	arguments, _ :=getString(args, "args")
	parts := strings.Fields(arguments)
	cmd := exec.CommandContext(ctx, command, parts...)
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("execution failed: %v", e))
}

	return ok(string(output))
}