package mcpimpl

import "context"

func HandleGetLayerInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	layerName, _ :=getString(args, "layer")
	if layerName == "" {
		return err("layer name is required")
}

	return ok("Layer info for " + layerName)
}

func HandleRunAlgorithm(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	algorithm, _ :=getString(args, "algorithm")
	params, _ :=getString(args, "params")
	if algorithm == "" {
		return err("algorithm is required")
}

	return success("Running algorithm " + algorithm + " with params " + params)
}