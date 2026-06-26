package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func HandleRunShell(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	out, e := exec.CommandContext(ctx, "sh", "-c", cmd).Output()
	if e != nil {
		return err(fmt.Sprintf("execution failed: %v", e))
}

	return ok(string(out))
}

func HandleGetEnv(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	val, found := os.LookupEnv(key)
	if !found {
		return err(fmt.Sprintf("env variable %q not set", key))
}

	return ok(val)
}