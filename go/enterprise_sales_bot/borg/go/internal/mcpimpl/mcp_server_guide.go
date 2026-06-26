package mcpimpl

import "context"

func HandleGetGuide_mcp_server_guide(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	var guide string
	switch topic {
	case "setup":
		guide = "Setup requires a Figma token and running the MCP server with the correct parameters."
	case "usage":
		guide = "Usage: send requests to the MCP server endpoint. Refer to the Figma API docs for available actions."
	case "authentication":
		guide = "Authentication: use a personal access token from Figma. Set it as an environment variable or config."
	default:
		guide = "Figma MCP Server Guide: covers setup, usage, and authentication. Use the 'topic' argument to get details."
	}
	return success(guide)
}

func HandleGetGuideTopics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("setup, usage, authentication")
}