package tools

import "context"

func HandleGetIntelligence(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("The Central Intelligence has determined that the answer is 42.")
}