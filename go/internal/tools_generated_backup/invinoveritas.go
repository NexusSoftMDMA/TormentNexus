package tools

import "context"

func HandleGetWineSuggestion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	grape, _ :=getString(args, "grape")
	if grape == "" {
		grape = "Cabernet Sauvignon"
	}
	return ok("I recommend a nice " + grape + " from Napa Valley.")
}