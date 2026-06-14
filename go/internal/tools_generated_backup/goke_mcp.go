package tools

import "context"

func HandleGenerateCliCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	toolName, _ :=getString(args, "tool_name")
	if toolName == "" {
		return err("tool_name is required")
}

	return success("mcp run " + toolName)
}

func HandleGetCliHelp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Usage: mcp run <tool_name>\nAvailable tools: generated dynamically from MCP server")
}