package mcpimpl

import (
	"context"
)

func HandleRustSdk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Rust SDK server is operational")
}

func HandleRustSdkGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok("Hello, " + name + "! Welcome to Rust SDK MCP server")
}