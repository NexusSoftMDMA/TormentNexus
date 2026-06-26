package mcpimpl

import (
	"context"
	"os/exec"
)

func HandleHermitExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	c := exec.CommandContext(ctx, "sh", "-c", cmd)
	out, e := c.CombinedOutput()
	if e != nil {
		return err("execution failed: " + e.Error() + "\n" + string(out))
}

	return ok(string(out))
}// touch 1781132121
