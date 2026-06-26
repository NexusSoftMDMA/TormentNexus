package mcpimpl

import (
	"context"
	"fmt"
)

// HandleReviewCode reviews the provided code.
func HandleReviewCode_open_code_review(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	language, _ :=getString(args, "language")
	if language == "" {
		language = "unknown"
	}
	review := fmt.Sprintf("Review of %s code:\n- Lines: %d\n- Suggestion: Ensure proper formatting and error handling.", language, len(code))
	return ok(review)
}