package mcpimpl

import "context"

func HandleNovyxCore(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return success("Novyx Core MCP server is running")
}