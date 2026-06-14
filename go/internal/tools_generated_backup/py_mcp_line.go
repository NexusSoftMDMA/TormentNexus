package tools

import "context"

// HandleX is a sample handler for the Py Mcp Line server.
func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		msg = "Hello from Py Mcp Line!"
	}
	return success(msg)
}