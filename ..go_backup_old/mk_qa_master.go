package tools

import "context"

func HandleAskQuestion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	question, _ :=getString(args, "question")
	return ok("Answer: " + question + " (from Mk Qa Master)")
}