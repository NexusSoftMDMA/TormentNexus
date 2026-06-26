package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func HandleRunMatlab(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	cmd := exec.CommandContext(ctx, "matlab", "-batch", code)
	out, e := cmd.Output()
	if e != nil {
		return err(fmt.Sprintf("failed to run matlab: %v", e))
}

	return success(string(out))
}

func HandleGetMatlabVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "matlab", "-version")
	out, e := cmd.Output()
	if e != nil {
		return err(fmt.Sprintf("failed to get matlab version: %v", e))
}

	return ok(strings.TrimSpace(string(out)))
}