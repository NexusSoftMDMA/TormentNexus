package mcpimpl

import "context"

func HandleGreet_model_context_protocol_tp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok("Hello, " + name)
}