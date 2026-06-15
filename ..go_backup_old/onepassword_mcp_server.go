package tools

import (
	"context"
	"fmt"
)

func HandleListItems(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	vaultID, _ :=getString(args, "vault_id")
	if vaultID == "" {
		return ok(`{"items": [{"id": "1", "title": "Test Item"}]}`)
	}
	return success(fmt.Sprintf(`{"items": [{"id": "1", "title": "Item in vault %s"}]}`, vaultID))
}

func HandleGetItem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	itemID, _ :=getString(args, "item_id")
	if itemID == "" {
		return err("item_id is required")
	}
	return success(fmt.Sprintf(`{"id": "%s", "title": "Item Details", "username": "user", "password": "pass"}`, itemID))
}