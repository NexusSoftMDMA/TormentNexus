package tools

import (
	"context"
)

func HandleGetPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ticker, _ :=getString(args, "ticker")
	if ticker == "" {
		return err("ticker is required")
}

	return success("Price of " + ticker + ": $100.00")
}

func HandleGetCompanyInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ticker, _ :=getString(args, "ticker")
	if ticker == "" {
		return err("ticker is required")
}

	return success("Company info for " + ticker + ": Morningstar Inc.")
}