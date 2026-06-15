package tools

import (
	"context"
	"os/exec"
)

func HandleRunCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	cmd := exec.CommandContext(ctx, "sh", "-c", code)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return err("execution error: " + e.Error())
}

	return ok(string(out))
}