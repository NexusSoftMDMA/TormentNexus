package mcpimpl

import "context"

func HandleGetLoops(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Available loops: loop_a, loop_b, loop_c")
}

func HandleGetLoopDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("Missing 'name' argument")
}

	return success("Details for loop " + name + ": data here")
}