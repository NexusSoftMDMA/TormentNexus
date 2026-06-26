package mcpimpl

import (
	"context"
	"os/exec"
	"strings"
)

func HandleTerminalExec(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmdStr, _ :=getString(args, "command")
	if cmdStr == "" {
		return err("command argument is required")
}

	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return err("empty command provided")
}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return err("execution failed: " + e.Error() + " | " + string(out))
	}
	return success(string(out))
}

func HandleTerminalConnect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverUrl, _ :=getString(args, "server_url")
	if serverUrl == "" {
		return err("server_url argument is required")
}

	cmd := exec.CommandContext(ctx, "npx", "-y", "@modelcontextprotocol/inspector", serverUrl)
	e := cmd.Start()
	if e != nil {
		return err("failed to start connection: " + e.Error())
	}
	return success("MCP server connection initiated for " + serverUrl)
}