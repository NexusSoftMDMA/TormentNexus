package tools

import "context"

func HandleGetAdvice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	advice := "Always be kind to yourself and others."
	if name != "" {
		advice = name + ", " + advice
	}
	return ok(advice)
}

func HandleGetMotivation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	motivation := "You are capable of amazing things!"
	return ok(motivation)
}