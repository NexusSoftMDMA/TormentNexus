package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

func HandleAnalyzeBinary(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	cmd := exec.CommandContext(ctx, "r2", "-q", "-c", "aaa; afl", path)
	var out bytes.Buffer
	cmd.Stdout = &out
	if e := cmd.Run(); e != nil {
		return err(fmt.Sprintf("failed to analyze binary: %v", e))
}

	return ok(out.String())
}

func HandleListFunctions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	cmd := exec.CommandContext(ctx, "r2", "-q", "-c", "afl", path)
	var out bytes.Buffer
	cmd.Stdout = &out
	if e := cmd.Run(); e != nil {
		return err(fmt.Sprintf("failed to list functions: %v", e))
}

	return ok(out.String())
}