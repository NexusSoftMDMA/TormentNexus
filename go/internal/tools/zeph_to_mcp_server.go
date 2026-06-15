//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
)

func HandleNotify(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	// In a real implementation, send notification via Zeph API
	return ok(fmt.Sprintf("Notification sent: %s", message))
}

func HandlePrompt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	promptName, _ :=getString(args, "name")
	userInput, _ :=getString(args, "input")
	if promptName == "" {
		return err("name is required")
}

	// In a real implementation, retrieve prompt from Zeph
	return ok(fmt.Sprintf("Prompt '%s' with input '%s'", promptName, userInput))
}package tools

import (
	"context"
	"fmt"
)