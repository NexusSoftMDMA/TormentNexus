package tools

import "context"

func HandleRunRoutine(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	routine, _ :=getString(args, "routine")
	if routine == "" {
		return err("routine is required")
}

	return ok("routine executed: " + routine)
}