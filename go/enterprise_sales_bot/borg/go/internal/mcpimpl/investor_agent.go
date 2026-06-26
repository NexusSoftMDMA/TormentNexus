package mcpimpl

import "context"

func HandleGetQuote_investor_agent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	return ok("Investor Agent: Current price for " + symbol + " is $150.00")
}