package mcpimpl

import (
	"context"
	"fmt"
)

func HandleGetVersion_perp_cli(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Perp Cli version 1.0.0")
}

func HandleGetHelp_perp_cli(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success(fmt.Sprintf("Usage: perp-cli [command] [options]\nAvailable commands: version, help"))
}