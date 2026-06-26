package mcpimpl

import "context"

func HandleDetectContextRot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	contextText, _ :=getString(args, "context")
	if len(contextText) > 500 {
		return success("Context rot detected: context is too long.")
}

	return success("No context rot detected.")
}