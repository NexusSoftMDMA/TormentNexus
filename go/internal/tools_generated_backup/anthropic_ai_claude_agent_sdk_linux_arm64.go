package tools

import (
	"context"
	"os/exec"
)

func HandleRunClaudeAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	input, _ :=getString(args, "input")
	cmd := exec.CommandContext(ctx, "claude-agent-sdk", command, input)
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err("failed to run claude agent: " + e.Error())
}

	return success(string(output))
}