package tools

import (
	"context"
	"fmt"
)

func HandleCodeReview(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("No code provided")
}

	language, _ :=getString(args, "language")
	review := fmt.Sprintf("Code review for %s:\n- Good structure\n- Consider adding comments\n- Check for edge cases", language)
	if language == "" {
		review = "Code review:\n- Good structure\n- Consider adding comments\n- Check for edge cases"
	}
	return ok(review)
}