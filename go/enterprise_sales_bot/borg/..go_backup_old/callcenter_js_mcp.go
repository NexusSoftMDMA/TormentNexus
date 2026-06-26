package tools

import (
	"context"
)

func HandleListCalls(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	status, _ :=getString(args, "status")
	return success(`[{"id":"1","status":"` + status + `"}]`)
}

func HandleGetCallDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	callId, _ :=getString(args, "callId")
	return success(`{"id":"` + callId + `","duration":120}`)
}