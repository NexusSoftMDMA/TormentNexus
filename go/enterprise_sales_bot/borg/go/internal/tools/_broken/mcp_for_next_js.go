package mcpimpl

import (
	"context"
)

func HandleNextJsInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Next.js App"
	}
	return success("Next.js project info for: " + name)
}

func HandleNextJsBuild(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("build command is required")
}

	return success("Running build: " + command)
}