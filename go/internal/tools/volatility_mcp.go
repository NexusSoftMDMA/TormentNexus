//go:build ignore
// +build ignore

package tools

import (
	"context"
)

func HandleListPlugins(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	plugins := []string{"imageinfo", "pslist", "psscan", "dlllist", "connections"	return ok(map[string]interface{}{
		"count":   len(plugins),
	}),
}

func HandleRunCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("missing required parameter: command")
	}
	profile, _ :=getString(args, "profile")
	result := "executed " + command + " with profile " + profile
	return success(result)
}
}