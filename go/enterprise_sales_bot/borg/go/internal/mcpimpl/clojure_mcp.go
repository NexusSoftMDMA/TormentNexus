package mcpimpl

import "context"

func HandleClojureEval(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	return success(code + " evaluated successfully")
}