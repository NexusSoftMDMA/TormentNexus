package mcpimpl

import "context"

func HandleListProjects_pm_copilot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = args
	return ok("Projects: [Project A, Project B]")
}

func HandleGetProject_pm_copilot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("missing project name")
}

	return ok("Project: " + name)
}