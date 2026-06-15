//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"os/exec"
)

func HandleGitStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		path = "."
	}
	out, e := exec.CommandContext(ctx, "git", "status", path).Output()
	if e != nil {
		return err(fmt.Sprintf("git status failed: %v", e))
}

	return ok(string(out))
}

func HandleGitLog(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		path = "."
	}
	maxCount, _ :=getString(args, "max_count")
	if maxCount == "" {
		maxCount = "10"
	}
	out, e := exec.CommandContext(ctx, "git", "log", "--oneline", "-"+maxCount, path).Output()
	if e != nil {
		return err(fmt.Sprintf("git log failed: %v", e))
}

	return ok(string(out))
}
