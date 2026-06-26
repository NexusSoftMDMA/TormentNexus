package mcpimpl

import "context"

func HandleX_modelcontextprotocolclients_jl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return success("Hello, MCP client! No name provided.")
}

	return success("Hello, " + name + "!")
}