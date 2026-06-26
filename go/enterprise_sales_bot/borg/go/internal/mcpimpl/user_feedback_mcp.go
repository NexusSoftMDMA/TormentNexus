package mcpimpl

import "context"

func HandleSubmitFeedback(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	return ok("Feedback submitted: " + message)
}

func HandleGetFeedback_user_feedback_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("No feedback available yet.")
}