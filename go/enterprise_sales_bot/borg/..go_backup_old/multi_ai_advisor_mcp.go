package tools

import (
	"context"
)

func HandleAsk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	question, _ :=getString(args, "question")
	if question == "" {
		return err("question is required")
}

	return ok("Advisor response: Consider multiple perspectives. " + question)
}

func HandleListAdvisors(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	advisors := []string{"AI Advisor 1", "AI Advisor 2", "AI Advisor 3"}
	return success(advisors)
}