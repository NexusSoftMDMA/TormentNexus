package tools

import (
	"context"
)

func HandleSearchVessels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	return success("Found vessels for query: " + query)
}

func HandleGetVessel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	imo, _ :=getString(args, "imo")
	if imo == "" {
		return err("imo is required")
}

	return success("Vessel details for IMO " + imo)
}