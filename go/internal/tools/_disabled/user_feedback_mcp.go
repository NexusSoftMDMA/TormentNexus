package tools

import "context"

func HandleSubmitFeedback(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	return ok("Feedback submitted: " + message)
}

func HandleGetFeedback(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("No feedback available yet.")
}