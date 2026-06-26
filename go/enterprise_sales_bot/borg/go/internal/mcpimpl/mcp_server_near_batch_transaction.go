package mcpimpl

import (
	"context"
	"fmt"
)

func HandleBatchTransaction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	signer, _ :=getString(args, "signer")
	if signer == "" {
		return err("signer is required")
}

	receiver, _ :=getString(args, "receiver")
	if receiver == "" {
		return err("receiver is required")
}

	actions := args["actions"]
	if actions == nil {
		return err("actions is required")
}

	actionsList, found := actions.([]interface{})
	if !found {
		return err("actions must be a list")
}

	return success(fmt.Sprintf("Batch transaction created from %s to %s with %d actions", signer, receiver, len(actionsList)))
}