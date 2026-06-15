package tools

import (
	"context"
	"strings"
)

func HandleValidateCommitMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("commit message is required")
}

	validPrefixes := []string{"feat:", "fix:", "chore:", "docs:", "test:"}
	found := false
	for _, p := range validPrefixes {
		if strings.HasPrefix(msg, p) {
			found = true
			break
		}
	}
	if !found {
		return err("commit message must start with one of: feat:, fix:, chore:, docs:, test:")
}

	return ok("commit message is valid")
}