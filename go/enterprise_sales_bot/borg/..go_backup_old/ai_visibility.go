package tools

import "context"

func HandleCheckVisibility(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	if model == "" {
		return err("model is required")
}

	return success("AI model " + model + " is visible")
}

func HandleListVisibleModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	visibleModels := []string{"gpt-4", "claude-3", "gemini-pro"}
	return ok("Visible models: " + strings.Join(visibleModels, ", "))
}