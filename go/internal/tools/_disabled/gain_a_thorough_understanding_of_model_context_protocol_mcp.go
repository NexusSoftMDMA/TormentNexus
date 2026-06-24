package tools

import "context"

func HandleGetMCPInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Gain A Thorough Understanding Of Model Context Protocol (MCP): MCP is an open protocol that standardizes how AI models interact with tools, data sources, and external systems. It enables structured request/response workflows and dynamic tool discovery.")
}