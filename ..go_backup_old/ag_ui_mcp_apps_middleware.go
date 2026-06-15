package tools

import (
	"context"
	"fmt"
)

func HandleListApps(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apps := []string{"app1", "app2", "app3"}
	return success(fmt.Sprintf("Available apps: %v", apps))
}

func HandleLaunchApp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("app name is required")
}

	return ok(fmt.Sprintf("Launched app: %s", name))
}