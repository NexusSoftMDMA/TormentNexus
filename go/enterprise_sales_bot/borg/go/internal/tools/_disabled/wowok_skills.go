package tools

import "context"

func HandleSkills(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("WoWok AI Skills: Use this tool to get guidance on correctly using WoWok MCP tools.")
}