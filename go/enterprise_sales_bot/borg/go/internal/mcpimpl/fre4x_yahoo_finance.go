package mcpimpl

import (
	"context"
	"fmt"
)

func HandleQuote_fre4x_yahoo_finance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	msg := fmt.Sprintf("Current price of %s is $150.00", symbol)
	return ok(msg)
}

func HandleProfile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	msg := fmt.Sprintf("Company profile for %s: Name: Example Corp, Sector: Technology", symbol)
	return ok(msg)
}// touch 1781132126
