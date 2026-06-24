package tools

import (
    "context"
)

func HandleGetProjectContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    name, _ :=getString(args, "name")
    return ok("Project context for: " + name)
}

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("List of projects: project1, project2, project3")
}