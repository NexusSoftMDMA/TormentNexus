package tools

import (
	"context"
	"fmt"
	"time"
)

var receipts []map[string]interface{}

func HandleCreateReceipt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	amount, _ :=getString(args, "amount")
	category, _ :=getString(args, "category")
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	receipt := map[string]interface{}{
		"id":       id,
		"title":    title,
		"amount":   amount,
		"category": category,
	}
	receipts = append(receipts, receipt)
	return ok("Receipt created: " + id)
}

func HandleListReceipts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	count := len(receipts)
	return ok(fmt.Sprintf("Total receipts: %d", count))
}