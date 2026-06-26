package tools

import (
	"fmt"
)

func HandleGetStockPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ticker, _ :=getString(args, "ticker")
	if ticker == "" {
		ticker = "AAPL"
	}
	price := 150.25
	return ok(fmt.Sprintf("Current price of %s is $%.2f", ticker, price))
}