package tools

import (
	"context"
	"fmt"
)

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filter, _ :=getString(args, "filter")
	if filter != "" {
		return ok(fmt.Sprintf("Projects matching: %s", filter))
}

	return ok("All projects")
}

func HandleGetProjectInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("project name is required")
}

	return ok(fmt.Sprintf("Project info for %s", name))
}