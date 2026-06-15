package tools

import (
	"context"
	"os/exec"
	"strings"
)

func HandleKnip(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dir, _ :=getString(args, "directory")
	cmd := exec.CommandContext(ctx, "knip")
	if dir != "" {
		cmd.Dir = dir
	}
	out, e := cmd.CombinedOutput()
	if e != nil {
		return err("knip failed: " + e.Error())
}

	return ok(strings.TrimSpace(string(out)))
}