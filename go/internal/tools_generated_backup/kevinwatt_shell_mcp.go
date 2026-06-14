package tools

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

func HandleShell(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	timeout, _ :=getInt(args, "timeout")
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()

	c := exec.CommandContext(ctx, "sh", "-c", cmd)
	out, e := c.CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("command failed: %s", e.Error()))
}

	return success(string(out))
}
}