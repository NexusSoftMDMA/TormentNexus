package tools

import (
	"context"
	"os/exec"
	"strings"
)

// HandleListDevices returns list of connected Android devices
func HandleListDevices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.Command("adb", "devices")
	out, e := cmd.Output()
	if e != nil {
		return err("failed to execute adb: " + e.Error())
}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var devices []string
	for _, line := range lines {
		if strings.Contains(line, "device") && !strings.Contains(line, "List of devices") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				devices = append(devices, parts[0])

		}
	}
	if len(devices) == 0 {
		return ok("No devices found")
}

	return ok("Devices: " + strings.Join(devices, ", "))
}

}

// HandleExecuteCommand executes an ADB command on a specific device
func HandleExecuteCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	device, _ :=getString(args, "device")
	command, _ :=getString(args, "command")
	if device == "" {
		return err("device is required")
}

	if command == "" {
		return err("command is required")
}

	cmd := exec.Command("adb", "-s", device, "shell", command)
	out, e := cmd.Output()
	if e != nil {
		return err("command failed: " + e.Error())
}

	return ok(string(out))
}