package mcpimpl

import (
	"context"
	"fmt"
)

func HandleGetBalance_zcash_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	msg := fmt.Sprintf(`{"address":"%s","balance":"0.0"}`, address)
	return ok(msg)
}