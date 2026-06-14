package tools

import "context"

func HandleAstronomyOracle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Current moon phase: Waxing Gibbous, 93% illumination.")
}