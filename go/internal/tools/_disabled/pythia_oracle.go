package tools

import "context"

func HandleAsk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	question, _ :=getString(args, "question")
	if question == "" {
		return err("no question provided")
}

	answer := "The oracle answers: " + question
	return ok(answer)
}