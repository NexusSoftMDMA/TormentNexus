package tools

import (
	"context"
)

func HandleGetWalletInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	return ok("Wallet address: " + address)
}