package mcpimpl

import (
	"context"
	"os/exec"
	"strings"
)

func HandleRun_trishchuk_codex_mcp_tool(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	timeout, _ :=getInt(args, "timeout")
	if timeout <= 0 {
		timeout = 30
	}
	c, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	out, e := exec.CommandContext(c, "sh", "-c", cmd).Output()
	if e != nil {
		return err("execution failed: " + e.Error())
}

	return ok(string(out))
}

func HandleList_trishchuk_codex_mcp_tool(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Returns a static list of available commands
	return ok("Available commands: run, list")
}