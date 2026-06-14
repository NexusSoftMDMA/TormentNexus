package tools

import (
	"context"
	"os/exec"
	"strings"
)

func HandleShellExec(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmdStr, _ :=getString(args, "command")
	if cmdStr == "" {
		return err("command argument is required")
	}
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return err("invalid command format")
	}
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return err("execution failed: " + e.Error() + " | " + string(out))
	}
	return success(string(out))
}

func HandleSessionInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	user, _ :=getString(args, "user")
	if host == "" {
		return err("host argument is required")
	}
	info := "Session ready"
	if user != "" {
		info = "Session ready for user " + user + " on " + host
	}
	return success(info)
}