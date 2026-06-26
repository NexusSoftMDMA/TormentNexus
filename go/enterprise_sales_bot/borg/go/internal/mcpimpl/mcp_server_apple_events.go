package mcpimpl

import (
	"context"
	"os/exec"
)

func HandleGetReminders(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.Command("osascript", "-e", `tell application "Reminders" to get name of reminders`)
	out, e := cmd.Output()
	if e != nil {
		return err("failed to get reminders: " + e.Error())
}

	return ok(string(out))
}

func HandleGetCalendarEvents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.Command("osascript", "-e", `tell application "Calendar" to get summary of events`)
	out, e := cmd.Output()
	if e != nil {
		return err("failed to get calendar events: " + e.Error())
}

	return ok(string(out))
}