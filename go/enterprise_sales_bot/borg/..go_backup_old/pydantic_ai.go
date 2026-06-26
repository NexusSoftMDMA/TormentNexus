package tools

import "context"

func HandlePydanticAiGenerate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	prompt, _ :=getString(args, "prompt")
	return success("Generated response for model " + model + ": " + prompt)
}

func HandlePydanticAiListModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success(`["gpt-4","gpt-3.5","claude-3"]`)
}