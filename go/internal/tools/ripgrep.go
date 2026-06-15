//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// HandleRipgrep implements the ripgrep tool natively with path sanitization.
func HandleRipgrep(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pattern, _ := getString(args, "pattern")
	if pattern == "" {
		return err("pattern is required")
	}

	path, _ := getString(args, "path")
	if path == "" {
		path = "."
	}

	// 1. Get absolute path of requested directory
	absPath, errAbs := filepath.Abs(path)
	if errAbs != nil {
		return err(fmt.Sprintf("invalid path: %v", errAbs))
	}

	// 2. Get absolute path of current working directory (workspace root)
	cwd, errWd := os.Getwd()
	if errWd != nil {
		return err("failed to get working directory")
	}

	// 3. Ensure the requested path is within the workspace boundary
	if !strings.HasPrefix(absPath, cwd) {
		return err("security violation: path is outside of workspace boundary")
	}

	cmd := exec.CommandContext(ctx, "rg", "--json", "--", pattern, absPath)
	output, e := cmd.CombinedOutput()
	if e != nil {
		if exitErr, okVal := e.(*exec.ExitError); okVal && exitErr.ExitCode() == 1 {
			return ok("[]")
		}
		return err(fmt.Sprintf("ripgrep failed: %v", e))
	}
	return ok(string(output))
}
