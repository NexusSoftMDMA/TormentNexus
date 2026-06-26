package tools

import (
	"context"
	"os/exec"
)

func HandleListSimulators(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	out, e := exec.CommandContext(ctx, "xcrun", "simctl", "list", "devices").Output()
	if e != nil {
		return err("failed to list simulators: " + e.Error())
}

	return success(string(out))
}