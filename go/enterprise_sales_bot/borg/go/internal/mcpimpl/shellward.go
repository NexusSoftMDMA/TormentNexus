package mcpimpl

import (
	"bytes"
	"context"
	"os/exec"
)

func HandleExecute_shellward(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command is required")
}

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	e := cmd.Run()
	if e != nil {
		return err("execution failed: " + e.Error() + "\nstderr: " + stderr.String())
}

	return ok("exit code 0\nstdout: " + stdout.String() + "\nstderr: " + stderr.String())
}