package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Helper to get device serial: args, env, then first device from adb devices
func getDeviceSerial(args map[string]interface{}) string {
	if s, found := args["serial"]; found {
		if str, ok2 := s.(string); ok2 && str != "" {
			return str
		}
	}
	if env := os.Getenv("ANDROID_MCP_DEVICE"); env != "" {
		return env
	}
	out, cmdErr := exec.Command("adb", "devices").Output()
	if cmdErr != nil {
		return ""
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == "device" {
			return fields[0]
		}
	}
	return ""
}

// runADB runs adb with optional -s device prefix and returns combined output
func runADB(args ...string) (string, error) {
	cmd := exec.Command("adb", args...)
	out, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return "", fmt.Errorf("adb %s failed: %v - %s", strings.Join(args, " "), cmdErr, string(out))
}

	return string(out), nil
}

// runADBWithDevice wraps runADB with -s device prefix if device is not empty
func runADBWithDevice(device string, args ...string) (string, error) {
	var fullArgs []string
	if device != "" {
		fullArgs = append(fullArgs, "-s", device)

	fullArgs = append(fullArgs, args...)
	return runADB(fullArgs...)
}

}

// mapButtonToKeycode maps common button names to Android keycodes
func mapButtonToKeycode(button string) string {
	button = strings.ToLower(strings.TrimSpace(button))
	switch button {
	case "back":
		return "KEYCODE_BACK"
}
	case "home":
		return "KEYCODE_HOME"
}
	case "menu":
		return "KEYCODE_MENU"
	case "volume_up":
		return "KEYCODE_VOLUME_UP"
	case "volume_down":
		return "KEYCODE_VOLUME_DOWN"
	case "power":
		return "KEYCODE_POWER"
	case "recent_apps", "recent":
		return "KEYCODE_APP_SWITCH"
	case "enter":
		return "KEYCODE_ENTER"
	case "delete":
		return "KEYCODE_DEL"
	case "space":
		return "KEYCODE_SPACE"
	default:
		return "KEYCODE_UNKNOWN"
	}
}

// HandleListDevices returns available ADB devices
func HandleListDevices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	out, cmdErr := runADB("devices")
	if cmdErr != nil {
		return err("failed to list devices: " + cmdErr.Error())

	return ok(fmt.Sprintf("ADB Devices:\n%s", out))

// HandleConnectDevice connects to a device via ADB
func HandleConnectDevice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serial, _ :=getString(args, "serial")
	if serial == "" {
		return err("missing required argument: serial")
}

	out, cmdErr := runADB("connect", serial)
	if cmdErr != nil {
		return err("connection failed: " + cmdErr.Error())
}

	return ok(fmt.Sprintf("Connect result: %s", out))
}

// HandleClick clicks at coordinates (x, y)
func HandleClick(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	x, _ :=getInt(args, "x")
	y, _ :=getInt(args, "y")
	device := getDeviceSerial(args)
	out, cmdErr := runADBWithDevice(device, "shell", "input", "tap", fmt.Sprintf("%d", x), fmt.Sprintf("%d", y))
	if cmdErr != nil {
		return err("click failed: " + cmdErr.Error())
}

	return ok(fmt.Sprintf("Clicked at (%d,%d)\n%s", x, y, out))
}

// HandleLongClick long-presses at (x, y) using swipe with duration
func HandleLongClick(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	x, _ :=getInt(args, "x")
	y, _ :=getInt(args, "y")
	device := getDeviceSerial(args)
	// long press simulated via swipe with 1000ms duration
	out, cmdErr := runADBWithDevice(device, "shell", "input", "swipe",
		fmt.Sprintf("%d", x), fmt.Sprintf("%d", y),
		fmt.Sprintf("%d", x), fmt.Sprintf("%d", y),
		"1000")
	if cmdErr != nil {
		return err("long click failed: " + cmdErr.Error())
}

	return ok(fmt.Sprintf("Long clicked at (%d,%d)\n%s", x, y, out))
}

// HandleSwipe performs swipe from (x1,y1) to (x2,y2)
func HandleSwipe(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	x1, _ :=getInt(args, "x1")
	y1, _ :=getInt(args, "y1")
	x2, _ :=getInt(args, "x2")
	y2, _ :=getInt(args, "y2")
	device := getDeviceSerial(args)
	out, cmdErr := runADBWithDevice(device, "shell", "input", "swipe",
		fmt.Sprintf("%d", x1), fmt.Sprintf("%d", y1),
		fmt.Sprintf("%d", x2), fmt.Sprintf("%d", y2))
	if cmdErr != nil {
		return err("swipe failed: " + cmdErr.Error())
}

	return ok(fmt.Sprintf("Swiped from (%d,%d) to (%d,%d)\n%s", x1, y1, x2, x2, out))
}