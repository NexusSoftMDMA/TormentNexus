package mcpimpl

import (
	"fmt"
)

func HandleGetStockPrice_profitspot_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ticker, _ :=getString(args, "ticker")
	if ticker == "" {
		ticker = "AAPL"
	}
	price := 150.25
	return ok(fmt.Sprintf("Current price of %s is $%.2f", ticker, price))
}