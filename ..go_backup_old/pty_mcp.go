package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func HandleExec(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmdStr, _ :=getString(args, "command")
	if cmdStr == "" {
		return err("missing command")
}

	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return err("invalid command")
}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("exec failed: %v - %s", e, string(output)))
}

	return ok(string(output))
}

func HandleList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	available := []string{"exec"}
	return success(fmt.Sprintf("available commands: %s", strings.Join(available, ", ")))
}