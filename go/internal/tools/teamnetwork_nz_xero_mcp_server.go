//go:build ignore
// +build ignore

package tools

import (
	"context"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	switch action {
	case "get_invoices":
		return success("Invoices: INV-001, INV-002")
	case "get_contacts":
		return success("Contacts: John Doe, Jane Smith")
	default:
		return err("unknown action: " + action)

}
}