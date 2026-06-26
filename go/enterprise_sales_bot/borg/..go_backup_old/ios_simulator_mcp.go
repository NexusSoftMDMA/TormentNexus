package tools

import (
	"context"
	"fmt"
	"os/exec"
)

func HandleListSimulators(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "xcrun", "simctl", "list", "--json", "devices")
	output, e := cmd.Output()
	if e != nil {
		return err(fmt.Sprintf("failed to list simulators: %v", e))
}

	return ok(string(output))
}

func HandleBootSimulator(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	udid, _ :=getString(args, "udid")
	if udid == "" {
		return err("udid is required")
}

	cmd := exec.CommandContext(ctx, "xcrun", "simctl", "boot", udid)
	if e := cmd.Run(); e != nil {
		return err(fmt.Sprintf("failed to boot simulator: %v", e))
}

	return success("Simulator booted successfully")
}