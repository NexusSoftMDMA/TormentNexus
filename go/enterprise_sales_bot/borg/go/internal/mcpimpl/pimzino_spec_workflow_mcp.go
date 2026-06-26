package mcpimpl

import (
	"context"
	"encoding/json"
)

func HandleSubmitSpec(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	spec, _ :=getString(args, "spec")
	if spec == "" {
		return err("spec is required")
}

	return ok("spec submitted: " + spec)
}

func HandleGetDashboard_pimzino_spec_workflow_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	data := map[string]string{"status": "active", "specs": "3"}
	jsonData, _ := json.Marshal(data)
	return ok(string(jsonData))
}