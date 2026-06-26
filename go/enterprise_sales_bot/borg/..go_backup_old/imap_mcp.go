package tools

import "context"

func HandleListMailboxes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Mailboxes: INBOX, Sent, Trash")
}

func HandleFetchEmail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = getString(args, "id")
	return ok("Email content: Placeholder")
}