package mcpimpl

import (
	"context"
	"os/exec"
	"time"
)

func HandleSandboxRun(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	timeout, _ :=getInt(args, "timeout")
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return err("exec failed: " + string(out) + ": " + e.Error())
}

	return ok("output: " + string(out))
}

}

func HandleSandboxList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("No sandboxes currently active")
}