package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func HandleGetStrings(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	file, _ :=getString(args, "file")
	if file == "" {
		return err("missing file path")
}

	out, e := exec.CommandContext(ctx, "strings", file).Output()
	if e != nil {
		return err(fmt.Sprintf("strings failed: %v", e))
}

	return ok(strings.TrimSpace(string(out)))
}

func HandleGetHash(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	file, _ :=getString(args, "file")
	if file == "" {
		return err("missing file path")
}

	out, e := exec.CommandContext(ctx, "sha256sum", file).Output()
	if e != nil {
		return err(fmt.Sprintf("sha256sum failed: %v", e))
}

	// output: hash  filename
	parts := strings.Fields(string(out))
	if len(parts) < 1 {
		return err("unexpected sha256sum output")
}

	return ok(parts[0])
}