package mcpimpl

import (
	"context"
	"os/exec"
)

func HandleRunCommand_mcp_server_bash_sdk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	out, e := exec.CommandContext(ctx, "bash", "-c", cmd).Output()
	if e != nil {
		return err(e.Error())
}

	return ok(string(out))
}

func HandleEcho_mcp_server_bash_sdk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}