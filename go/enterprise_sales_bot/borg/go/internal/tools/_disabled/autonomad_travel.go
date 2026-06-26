package tools

import "context"

func HandleGetDestinations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	continent, _ :=getString(args, "continent")
	data := "Destinations in " + continent + ": Paris, Tokyo, New York"
	return success(data)
}

func HandleGetTravelAdvice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	country, _ :=getString(args, "country")
	advice := "Travel advisory for " + country + ": Exercise normal precautions."
	return success(advice)
}