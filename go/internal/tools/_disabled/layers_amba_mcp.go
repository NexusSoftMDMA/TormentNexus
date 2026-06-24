package tools

import (
	"context"
	"fmt"
)

func HandleAmbaGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "world"
	}
	msg := fmt.Sprintf("Hello, %s! Welcome to amba.", name)
	return ok(msg)
}

func HandleAmbaProvision(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project_name")
	if project == "" {
		return err("project_name is required")
}

	result := fmt.Sprintf("Project '%s' provisioned successfully.", project)
	return success(result)
}