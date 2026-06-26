package mcpimpl

import "context"

func HandleListMemories_memorylane(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userId, _ :=getString(args, "userId")
	return success("ListMemories for userId: " + userId)
}

func HandleGetMemory_memorylane(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	memoryId, _ :=getString(args, "memoryId")
	return ok("Memory " + memoryId + " retrieved successfully")
}