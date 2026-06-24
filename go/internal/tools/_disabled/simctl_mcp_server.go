package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

func HandleListDevices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "xcrun", "simctl", "list")
	var out bytes.Buffer
	cmd.Stdout = &out
	if e := cmd.Run(); e != nil {
		return err("Failed to list devices: " + e.Error())
}

	return ok(out.String())
}

func HandleBootDevice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	udid, _ :=getString(args, "udid")
	if udid == "" {
		return err("Missing 'udid' argument")
}

	cmd := exec.CommandContext(ctx, "xcrun", "simctl", "boot", udid)
	if e := cmd.Run(); e != nil {
		return err("Failed to boot device: " + e.Error())
}

	return ok("Device booted successfully")
}