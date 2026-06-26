package tools

import "context"

func HandlePrism(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return success("Prism Mcp says hello")
}

	return success("Prism Mcp: " + message)
}