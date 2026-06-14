package tools

import "context"

func HandleMcpCreate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	desc, _ :=getString(args, "description")
	if name == "" {
		return err("name is required")
}

	msg := "Created resource: " + name
	if desc != "" {
		msg += " (" + desc + ")"
	}
	return ok(msg)
}