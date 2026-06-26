package mcpimpl

import (
	"context"
	"os/exec"
)

func HandleAnalyzeBinary_ida_headless_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	binaryPath, _ :=getString(args, "binary_path")
	if binaryPath == "" {
		return err("binary_path is required")
}

	cmd := exec.CommandContext(ctx, "ida64", "-A", "-L/dev/stdout", binaryPath)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return err("analysis failed: " + e.Error())
}

	return success("Analysis result:\n" + string(out))
}