//go:build ignore
// +build ignore

package tools

import "context"

func HandleCreateReceipt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	amount, _ :=getString(args, "amount")
	if name == "" {
		return err("name is required")
}

	return success("receipt created: " + name)
}

func HandleListReceipts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("list of receipts placeholder")
}
