package tools

import (
	"context"
	"os/exec"
)

func HandleDecompile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	cmd := exec.CommandContext(ctx, "javap", "-p", "-c", path)
	out, e := cmd.Output()
	if e != nil {
		return err("decompilation failed: " + e.Error())
}

	return ok("decompiled:\n" + string(out))
}