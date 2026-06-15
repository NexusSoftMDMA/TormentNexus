package tools

import (
	"context"
	"os/exec"
)

func HandleOpenURL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	cmd := exec.Command("osascript", "-e", `tell application "Safari" to open location "`+url+`"`)
	if e := cmd.Run(); e != nil {
		return err("failed to open URL: " + e.Error())
}

	return ok("URL opened")
}

func HandleGetTabs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.Command("osascript", "-e", `tell application "Safari" to get name of every tab of every window`)
	output, e := cmd.Output()
	if e != nil {
		return err("failed to get tabs: " + e.Error())
}

	return success(string(output))
}