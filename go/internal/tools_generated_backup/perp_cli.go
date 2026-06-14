package tools

import (
	"context"
	"fmt"
)

func HandleGetVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Perp Cli version 1.0.0")
}

func HandleGetHelp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success(fmt.Sprintf("Usage: perp-cli [command] [options]\nAvailable commands: version, help"))
}