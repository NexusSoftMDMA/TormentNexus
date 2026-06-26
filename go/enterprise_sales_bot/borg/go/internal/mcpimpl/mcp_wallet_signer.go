package mcpimpl

import "context"

func HandleSignMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	signature := "0x" + message
	return ok(map[string]interface{}{"signature": signature})
}

func HandleGetAddress(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address := "0x1234567890abcdef"
	return ok(map[string]interface{}{"address": address})
}