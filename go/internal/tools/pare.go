//go:build ignore
// +build ignore

package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

func HandleGitStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dir, _ :=getString(args, "path")
	if dir == "" {
		dir = "."
	}
	cmd := exec.CommandContext(ctx, "git", "status")
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if e := cmd.Run(); e != nil {
		return err(fmt.Sprintf("git status failed: %v", e))
}

	return success(out.String())
}

func HandleNpmRun(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	script, _ :=getString(args, "script")
	if script == "" {
		return err("missing required argument: script")
}

	dir, _ :=getString(args, "path")
	if dir == "" {
		dir = "."
	}
	cmd := exec.CommandContext(ctx, "npm", "run", script)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if e := cmd.Run(); e != nil {
		return err(fmt.Sprintf("npm run failed: %v", e))
}

	return success(out.String())
}
