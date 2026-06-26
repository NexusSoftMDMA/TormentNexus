package mcpimpl

import (
	"context"
)

func HandleGetProjectInfo_davinci_resolve_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "project_name")
	if name == "" {
		name = "Untitled Project"
	}
	return ok("Project: " + name)
}

func HandleSetTimelineColor(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	color, _ :=getString(args, "color")
	if color == "" {
		return err("color is required")
}

	_ = getString(args, "project")
	return success("Timeline color set to " + color)
}