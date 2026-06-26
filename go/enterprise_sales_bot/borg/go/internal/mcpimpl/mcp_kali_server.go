package mcpimpl

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

func HandleRunCommand_mcp_kali_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	c := exec.CommandContext(ctx, "sh", "-c", cmd)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	e := c.Run()
	if e != nil {
		return err(fmt.Sprintf("command failed: %s; stderr: %s", e.Error(), stderr.String()))
}

	return ok(fmt.Sprintf("stdout: %s\nstderr: %s", stdout.String(), stderr.String()))
}

func HandleSystemInfo_mcp_kali_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	c := exec.CommandContext(ctx, "uname", "-a")
	var out bytes.Buffer
	c.Stdout = &out
	e := c.Run()
	if e != nil {
		return err(fmt.Sprintf("failed to get system info: %s", e.Error()))
}

	return ok(fmt.Sprintf("System info: %s", out.String()))
}