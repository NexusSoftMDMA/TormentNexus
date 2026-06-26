package mcpimpl

import "context"

func HandleRedirect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("This is a redirect stub. Please use @signatrust/mcp-server instead.")
}