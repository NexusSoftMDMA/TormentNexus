package tools

import "context"

func HandleInspectTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Inspector is ready – tools list available on request.")
}

func HandleInspectResources(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Inspector is ready – resources list available on request.")
}