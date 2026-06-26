package mcpimpl

import "context"

func HandleGetContext_context_kit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	msg := "Context Kit received query: " + query
	return ok(msg)
}

func HandleListContexts_context_kit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Available contexts: default, user, session")
}