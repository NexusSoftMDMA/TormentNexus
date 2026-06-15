//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
)

func HandleListApps(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Available apps: app1, app2, app3")
}

func HandleBuildApp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectName, _ :=getString(args, "project_name")
	language, _ :=getString(args, "language")
	msg := fmt.Sprintf("Building app '%s' with language '%s'", projectName, language)
	return success(msg)
}
