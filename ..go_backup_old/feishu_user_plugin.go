package tools

import "context"

func HandleGetUserInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ :=getString(args, "user_id")
	return success("User info for: " + userID)
}