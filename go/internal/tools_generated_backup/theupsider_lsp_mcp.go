package tools

import "context"

func HandleLspRequest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	method, _ :=getString(args, "method")
	if method == "" {
		return err("method is required")
}

	return ok("request: " + method)
}

func HandleLspShutdown(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("lsp shutdown acknowledged")
}