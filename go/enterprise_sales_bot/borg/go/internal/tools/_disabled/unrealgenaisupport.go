package tools

import "context"

func HandleChat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	return ok("Chat received: " + prompt)
}

func HandleGenerate3D(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	return ok("Generate 3D using " + model)
}