package tools

import "context"

func HandleCoconutEval(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	return ok("Evaluated: " + code)
}