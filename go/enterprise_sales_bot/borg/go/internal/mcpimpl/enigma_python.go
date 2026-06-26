package mcpimpl

import (
	"context"
	"os/exec"
)

func HandleExecutePython_enigma_python(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code argument is required")
}

	cmd := exec.CommandContext(ctx, "python3", "-c", code)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return err("execution failed: " + e.Error())
}

	return ok(string(out))
}