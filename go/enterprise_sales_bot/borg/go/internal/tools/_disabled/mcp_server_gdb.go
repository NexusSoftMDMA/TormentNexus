package tools

import (
	"context"
	"os/exec"
	"strings"
)

func HandleGdb(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command parameter is required")
}

	out, e := exec.CommandContext(ctx, "gdb", "-batch", "-ex", cmd).CombinedOutput()
	if e != nil {
		return err("gdb failed: " + e.Error() + ": " + string(out))
}

	return ok(strings.TrimSpace(string(out)))
}