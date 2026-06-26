package tools

import (
	"context"
	"os/exec"
)

func HandleClaudeAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	cmd := exec.CommandContext(ctx, "claude-agent-sdk-linux-arm64-musl", prompt)
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err("execution failed: " + e.Error())
}

	return ok(string(output))
}