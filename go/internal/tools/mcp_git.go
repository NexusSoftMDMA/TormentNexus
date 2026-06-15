//go:build ignore
// +build ignore

package tools

import (
	"context"
	"os/exec"
	"strings"
)

func HandleGitStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--short")
	out, e := cmd.Output()
	if e != nil {
		return err("git status failed: " + e.Error())
}

	return ok(strings.TrimSpace(string(out)))
}

func HandleGitBranches(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "git", "branch")
	out, e := cmd.Output()
	if e != nil {
		return err("git branch failed: " + e.Error())
}

	return ok(strings.TrimSpace(string(out)))
}
