package tools

import (
	"context"
	"fmt"
	"os/exec"
)

func HandleSshExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	user, _ :=getString(args, "user")
	command, _ :=getString(args, "command")
	if host == "" || user == "" || command == "" {
		return err("host, user, and command are required")
}

	target := fmt.Sprintf("%s@%s", user, host)
	cmd := exec.CommandContext(ctx, "ssh", "-o", "StrictHostKeyChecking=no", target, command)
	output, e := cmd.Output()
	if e != nil {
		return err(fmt.Sprintf("SSH execution failed: %v", e))
}

	return ok(string(output))
}