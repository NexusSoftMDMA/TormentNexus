package tools

import "context"

func HandleSolve(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	problem, _ :=getString(args, "problem")
	if problem == "" {
		return err("problem is required")
}

	result := "Solved: " + problem
	return success(result)
}