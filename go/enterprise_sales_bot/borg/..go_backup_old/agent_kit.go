package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

func HandleInbox(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	inbox := []map[string]string{
		{"from": "alice@example.com", "subject": "Hello", "body": "Hi there!"},
		{"from": "bob@example.com", "subject": "Meeting", "body": "Reminder for tomorrow."},
	}
	data, e := json.Marshal(inbox)
	if e != nil {
		return err("failed to marshal inbox")
}

	return success(string(data))
}

func HandleSend(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	msg := fmt.Sprintf("Email sent to %s with subject '%s'", to, subject)
	return ok(msg)
}