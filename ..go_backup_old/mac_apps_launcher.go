package tools

import (
	"context"
	"os/exec"
)

func HandleLaunchApp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	app, _ :=getString(args, "app")
	if app == "" {
		return err("app name is required")
}

	cmd := exec.Command("open", "-a", app)
	e := cmd.Run()
	if e != nil {
		return err("failed to launch " + app + ": " + e.Error())
}

	return ok("launched " + app)
}