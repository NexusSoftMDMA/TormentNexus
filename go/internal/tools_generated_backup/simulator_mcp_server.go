package tools

import "context"

func HandleSimulate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Simulation completed successfully")
}

func HandleGetSimulation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	simType, _ :=getString(args, "type")
	if simType == "" {
		return success("No simulation type specified")
}

	return success("Simulation of type: " + simType)
}