package tools

import "context"

func HandleGetInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = ctx
	_ = args
	return ok("Model Context Protocol provides a framework for connecting AI models with external tools and data sources securely. Example: an MCP server can expose a database query tool that the AI calls via JSON-RPC.")
}

func HandleSecure(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = ctx
	_ = args
	return ok("To secure an MCP server: use API keys, HTTPS, input validation, rate limiting, and allow-listed tools. Example: reject requests without a valid Authorization header.")
}