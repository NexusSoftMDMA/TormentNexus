package tools

import "context"

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Aiven projects listed")
}

func HandleGetProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "project_name")
	return success("Aiven project: " + name)
}