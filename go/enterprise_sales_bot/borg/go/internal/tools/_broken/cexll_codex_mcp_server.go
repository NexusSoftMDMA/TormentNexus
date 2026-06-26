package mcpimpl

import (
	"context"
	"os"
	"os/exec"
	"strings"
)

func HandleRunShell(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := strings.TrimSpace(getString(args, "command"))
	if cmd == "" {
		return err("command is required")
}

	out, e := exec.CommandContext(ctx, "sh", "-c", cmd).CombinedOutput()
	if e != nil {
		return err("execution failed: " + e.Error() + "\n" + string(out))
}

	return success(string(out))
}

func HandleListDir_cexll_codex_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		path = "."
	}
	entries, e := os.ReadDir(path)
	if e != nil {
		return err("cannot read directory: " + e.Error())
}

	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())

	return success(strings.Join(names, "\n"))
}
}