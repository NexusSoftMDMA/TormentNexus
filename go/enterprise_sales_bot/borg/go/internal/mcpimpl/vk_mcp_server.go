package mcpimpl

import "context"

func HandlePing_vk_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("VK MCP Server is running")
}