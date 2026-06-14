package tools

import "context"

func HandleCodeCall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	script, _ :=getString(args, "script")
	language, _ :=getString(args, "language")
	result := "Executed " + language + " script: " + script
	return ok(result)
}