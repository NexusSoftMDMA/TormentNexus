//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

func HandleBash(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmdStr, _ :=getString(args, "command")
	if cmdStr == "" {
		return err("command is required")
}

	timeoutSec, _ :=getInt(args, "timeout")
	var cancel context.CancelFunc
	if timeoutSec > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
		defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", cmdStr)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("command failed: %s\n%s", e.Error(), string(out)))
}

	return success(string(out))
}
}
