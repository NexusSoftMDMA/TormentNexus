package mcpimpl

import (
	"context"
	"os/exec"
	"strings"
)

func HandleAdbDevices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "adb", "devices")
	output, e := cmd.Output()
	if e != nil {
		return err("failed to execute adb devices: " + e.Error())
}

	return success(string(output))
}

func HandleAdbExec(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command argument is required")
}

	parts := strings.Fields(command)
	if len(parts) == 0 {
		return err("invalid command")
}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err("adb command failed: " + e.Error() + "\n" + string(output))
}

	return success(string(output))
}