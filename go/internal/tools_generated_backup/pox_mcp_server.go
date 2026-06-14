package tools

import (
	"context"
	"fmt"
)

func HandlePoxInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "Pox"
	}
	msg := fmt.Sprintf("Hello from %s Mcp Server", name)
	return ok(msg)
}