package mcpimpl

import (
	"context"
	"os/exec"
	"strings"
)

func HandleGetFrontmostApp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "osascript", "-e", `tell application "System Events" to get name of first application process whose frontmost is true`)
	out, e := cmd.Output()
	if e != nil {
		return err("failed to get frontmost app: " + e.Error())
}

	name := strings.TrimSpace(string(out))
	if name == "" {
		return err("no frontmost app found")
}

	return success(name)
}

func HandleExecuteAppleScript(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	script, _ :=getString(args, "script")
	if script == "" {
		return err("script argument is required")
}

	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return err("AppleScript error: " + e.Error() + ": " + string(out))
}

	return ok(strings.TrimSpace(string(out)))
}