package mcpimpl

import "context"

func HandleBuild_forge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	if project == "" {
		return err("project is required")
}

	return success("built project " + project)
}