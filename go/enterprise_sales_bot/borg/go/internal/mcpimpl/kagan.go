package mcpimpl

import "context"

func HandleKagan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return ok("Hello from Kagan!")
}

	return ok("Hello, " + name + "!")
}