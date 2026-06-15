package tools

import "context"

func HandleOpenProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	return ok("Opened project: " + project)
}

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Listed projects")
}