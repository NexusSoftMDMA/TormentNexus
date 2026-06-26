package mcpimpl

import "context"

func HandlePing_uniswap_trader_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}

func HandleGetQuote_uniswap_trader_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	getString(args, "tokenIn")
	return ok("quote retrieved")
}