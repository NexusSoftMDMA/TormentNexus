package tools

import "context"

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = args
	return ok("Projects: [Project A, Project B]")
}

func HandleGetProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("missing project name")
}

	return ok("Project: " + name)
}