package tools

import (
	"context"
)

func HandleGetPersonalization(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ :=getString(args, "userId")
	if userID == "" {
		return err("userId is required")
}

	return success("personalization data for user " + userID)
}

func HandleSetPersonalization(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ :=getString(args, "userId")
	if userID == "" {
		return err("userId is required")
}

	prefs, _ :=getString(args, "preferences")
	_ = prefs // preferences ignored for brevity
	return ok("personalization updated for user " + userID)
}