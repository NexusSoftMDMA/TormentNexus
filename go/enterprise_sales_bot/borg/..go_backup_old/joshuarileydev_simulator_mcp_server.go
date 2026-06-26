package tools

import (
	"context"
	"os/exec"
	"strings"
)

func HandleListSimulators(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "xcrun", "simctl", "list", "devices", "available")
	out, e := cmd.Output()
	if e != nil {
		return err("failed to list simulators: " + e.Error())
	}
	return success(strings.TrimSpace(string(out)))
}

func HandleBootSimulator(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	udid, _ :=getString(args, "udid")
	if udid == "" {
		return err("udid is required")
	}
	cmd := exec.CommandContext(ctx, "xcrun", "simctl", "boot", udid)
	e := cmd.Run()
	if e != nil {
		return err("failed to boot simulator: " + e.Error())
	}
	return success("simulator " + udid + " booted successfully")
}// touch 1781132128
