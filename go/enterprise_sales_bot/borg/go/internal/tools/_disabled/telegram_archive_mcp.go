package tools

import (
	"context"
)

func HandleListArchives(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filter, _ :=getString(args, "filter")
	if filter == "" {
		filter = "all"
	}
	return success("Archives list for filter: " + filter)
}

func HandleGetArchiveMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	chatID, _ :=getString(args, "chat_id")
	messageID, _ :=getString(args, "message_id")
	if chatID == "" || messageID == "" {
		return err("chat_id and message_id required")
}

	return success("Message " + messageID + " from chat " + chatID)
}