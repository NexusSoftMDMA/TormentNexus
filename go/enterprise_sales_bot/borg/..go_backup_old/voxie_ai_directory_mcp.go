package tools

import "context"

func HandleGetDirectory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")
	if category == "" {
		return ok("Please specify a category.")
}

	return success("Directory listing for category: " + category)
}