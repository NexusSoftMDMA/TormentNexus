package tools

import "context"

func HandleAgentforge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return success("Hello from Agentforge, " + name)
}