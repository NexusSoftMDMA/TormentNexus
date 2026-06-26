package tools

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

func HandleListDevices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "adb", "devices")
	var out bytes.Buffer
	cmd.Stdout = &out
	if e := cmd.Run(); e != nil {
		return err("failed to list devices: " + e.Error())
}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	var devices []string
	for _, line := range lines {
		if strings.HasSuffix(line, "\tdevice") {
			devices = append(devices, strings.Fields(line)[0])

	}
	return ok("devices: " + strings.Join(devices, ", "))
}

}

func HandleScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	device, _ :=getString(args, "device_id")
	if device == "" {
		return err("device_id is required")
}

	shellCmd := exec.CommandContext(ctx, "adb", "-s", device, "shell", "screencap", "/sdcard/screenshot.png")
	if e := shellCmd.Run(); e != nil {
		return err("screencap failed: " + e.Error())
}

	pullCmd := exec.CommandContext(ctx, "adb", "-s", device, "pull", "/sdcard/screenshot.png", "/tmp/screenshot.png")
	if e := pullCmd.Run(); e != nil {
		return err("pull failed: " + e.Error())
}

	return success("screenshot saved to /tmp/screenshot.png")
}