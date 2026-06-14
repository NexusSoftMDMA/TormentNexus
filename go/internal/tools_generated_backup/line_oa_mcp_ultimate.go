package tools

import "context"

func HandleSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	return success("Message sent: " + message)
}

func HandleGetProfile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ :=getString(args, "userId")
	if userID == "" {
		return err("userId is required")
}

	return ok("Profile retrieved for " + userID)
}