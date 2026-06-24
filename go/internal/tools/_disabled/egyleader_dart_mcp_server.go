package tools

import (
	"context"
	"os/exec"
)

func HandleDartFormat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	cmd := exec.CommandContext(ctx, "dart", "format", path)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return err("format failed: " + string(out))
}

	return ok(string(out))
}

func HandleDartAnalyze(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	cmd := exec.CommandContext(ctx, "dart", "analyze", path)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return err("analyze failed: " + string(out))
}

	return ok(string(out))
}