//go:build ignore
// +build ignore

package tools

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

func HandleVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "claude-code", "--version")
	out := new(bytes.Buffer)
	cmd.Stdout = out
	cmd.Stderr = new(bytes.Buffer)
	e := cmd.Run()
	if e != nil {
		return err("failed to run claude-code: " + e.Error())
}

	return ok("claude-code version: " + strings.TrimSpace(out.String()))
}

func HandleExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmdline, _ :=getString(args, "commandline")
	if cmdline == "" {
		return err("commandline argument is required")
}

	parts := strings.Fields(cmdline)
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	out := new(bytes.Buffer)
	cmd.Stdout = out
	cmd.Stderr = new(bytes.Buffer)
	e := cmd.Run()
	if e != nil {
		return err("execution failed: " + e.Error())
}

	return ok("output: " + strings.TrimSpace(out.String()))
}
