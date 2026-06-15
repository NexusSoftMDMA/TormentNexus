//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"os/exec"
)

func HandleRipgrep(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pattern, _ :=getString(args, "pattern")
	if pattern == "" {
		return err("pattern is required")
}

	path, _ :=getString(args, "path")
	if path == "" {
		path = "."
	}
	caseSensitive, _ :=getBool(args, "caseSensitive")
	rgArgs := []string{"--color", "never", pattern, path}
	if !caseSensitive {
		rgArgs = append([]string{"-i"}, rgArgs...)

	cmd := exec.CommandContext(ctx, "rg", rgArgs...)
	output, e := cmd.CombinedOutput()
	if e != nil {
		if exitErr, found := e.(*exec.ExitError); found && exitErr.ExitCode() == 1 {
			return ok(string(output))
}

		return err(fmt.Sprintf("ripgrep failed: %v", e))
}

	return ok(string(output))
}
}
