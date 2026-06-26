package mcpimpl

import (
	"context"
)

func HandleGetPrice_morningstar_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ticker, _ :=getString(args, "ticker")
	if ticker == "" {
		return err("ticker is required")
}

	return success("Price of " + ticker + ": $100.00")
}

func HandleGetCompanyInfo_morningstar_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ticker, _ :=getString(args, "ticker")
	if ticker == "" {
		return err("ticker is required")
}

	return success("Company info for " + ticker + ": Morningstar Inc.")
}