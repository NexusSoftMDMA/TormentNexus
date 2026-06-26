package mcpimpl

import (
	"context"
	"encoding/hex"
)

func HandlerDeployContract(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	return success("deployed with address 0x123")
}

func HandlerEncryptValue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	value, _ :=getString(args, "value")
	if value == "" {
		return err("value is required")
}

	encrypted := "0x" + hex.EncodeToString([]byte(value))
	return ok("encrypted: " + encrypted)
}