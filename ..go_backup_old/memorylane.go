package tools

import "context"

func HandleListMemories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userId, _ :=getString(args, "userId")
	return success("ListMemories for userId: " + userId)
}

func HandleGetMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	memoryId, _ :=getString(args, "memoryId")
	return ok("Memory " + memoryId + " retrieved successfully")
}