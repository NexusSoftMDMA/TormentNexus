package tools

import (
	"context"
)

func HandleBossAgentCli(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command is required")
}

	return ok("Executed boss agent command: " + command)
}