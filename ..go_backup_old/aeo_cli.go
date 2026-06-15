package tools

import (
	"context"
	"os/exec"
)

func HandleExecuteCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command argument is required")
}

	shell, _ :=getString(args, "shell")
	if shell == "" {
		shell = "sh"
	}
	flag, _ :=getString(args, "flag")
	if flag == "" {
		flag = "-c"
	}
	output, e := exec.Command(shell, flag, cmd).Output()
	if e != nil {
		return err("execution failed: " + e.Error())
}

	return ok(string(output))
}

func HandleGetVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Aeo Cli MCP Server v1.0")
}