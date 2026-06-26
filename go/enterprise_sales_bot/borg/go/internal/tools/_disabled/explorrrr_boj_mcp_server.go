package tools

import (
	"context"
	"os/exec"
)

func HandleBojRun(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	out, e := exec.Command("npx", "@explorrrr/boj-mcp-server", cmd).CombinedOutput()
	if e != nil {
		return err("failed: " + e.Error())
}

	return success(string(out))
}

func HandleBojPing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}