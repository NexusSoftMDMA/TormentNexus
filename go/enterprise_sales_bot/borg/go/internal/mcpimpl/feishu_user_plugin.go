package mcpimpl

import "context"

func HandleGetUserInfo_feishu_user_plugin(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ :=getString(args, "user_id")
	return success("User info for: " + userID)
}