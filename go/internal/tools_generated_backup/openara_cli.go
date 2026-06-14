package tools

import (
	"context"
	"os/exec"
)

func HandleExecuteCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command is required")
}

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err("execution failed: " + e.Error())
}

	return ok(string(output))
}

func HandleOpenApplication(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	app, _ :=getString(args, "application")
	if app == "" {
		return err("application name is required")
}

	cmd := exec.CommandContext(ctx, "open", "-a", app)
	e := cmd.Run()
	if e != nil {
		return err("failed to open application: " + e.Error())
}

	return ok("opened " + app)
}