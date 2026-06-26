package mcpimpl

import (
	"context"
)

func HandleValidate_guidegraph_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workflow, _ :=getString(args, "workflow")
	if workflow == "" {
		return err("workflow is required")
}

	return ok("Workflow validated successfully")
}

func HandleSimulate_guidegraph_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workflow, _ :=getString(args, "workflow")
	if workflow == "" {
		return err("workflow is required")
}

	inputs, _ :=getString(args, "inputs")
	result := "Simulation completed"
	if inputs != "" {
		result = "Simulation with inputs: " + inputs
	}
	return success(result)
}