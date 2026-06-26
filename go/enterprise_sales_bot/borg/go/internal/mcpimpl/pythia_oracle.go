package mcpimpl

import "context"

func HandleAsk_pythia_oracle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	question, _ :=getString(args, "question")
	if question == "" {
		return err("no question provided")
}

	answer := "The oracle answers: " + question
	return ok(answer)
}