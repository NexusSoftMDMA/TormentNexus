package tools

import "context"

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	diagram, _ :=getString(args, "diagram")
	return ok("Processed diagram: " + diagram)
}