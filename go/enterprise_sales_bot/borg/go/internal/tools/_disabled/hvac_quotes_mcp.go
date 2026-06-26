package tools

import (
	"context"
	"fmt"
)

func HandleGetQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	system, _ :=getString(args, "system")
	size, _ :=getInt(args, "size")
	price := size * 100
	msg := fmt.Sprintf("Quote for %s system size %d: $%d", system, size, price)
	return success(msg)
}

func HandleListSystems(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	systems := []string{"split", "packaged", "mini-split", "heat-pump"}
	return ok(fmt.Sprintf("Available systems: %v", systems))
}