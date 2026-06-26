package mcpimpl

import (
	"context"
	"os/exec"
)

func HandleExecCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	container, _ :=getString(args, "container")
	command, _ :=getString(args, "command")
	if container == "" {
		return err("container is required")
}

	if command == "" {
		return err("command is required")
}

	cmd := exec.CommandContext(ctx, "docker", "exec", container, "sh", "-c", command)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return err("execution failed: " + string(out) + " - " + e.Error())
}

	return success(string(out))
}