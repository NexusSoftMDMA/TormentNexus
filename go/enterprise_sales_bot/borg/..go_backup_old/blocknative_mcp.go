package tools

import (
	"context"
	"fmt"
)

func HandleBlocknativeTransaction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	txHash, _ :=getString(args, "transactionHash")
	if txHash == "" {
		return err("transactionHash is required")
}

	return ok(fmt.Sprintf("Blocknative transaction status for %s: pending", txHash))
}

func HandleBlocknativeMonitorAddress(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	network, _ :=getString(args, "network")
	if network == "" {
		network = "ethereum"
	}
	return ok(fmt.Sprintf("Monitoring address %s on %s with Blocknative", address, network))
}