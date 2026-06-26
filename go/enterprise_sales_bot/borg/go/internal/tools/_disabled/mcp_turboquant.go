package tools

import (
	"context"
	"fmt"
)

func HandleCalculatePnl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	buyPrice, _ :=getInt(args, "buyPrice")
	sellPrice, _ :=getInt(args, "sellPrice")
	quantity, _ :=getInt(args, "quantity")
	profit := (sellPrice - buyPrice) * quantity
	return ok(fmt.Sprintf("Profit: %d", profit))
}