package tools

import "context"

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}

func HandleGetQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	getString(args, "tokenIn")
	return ok("quote retrieved")
}