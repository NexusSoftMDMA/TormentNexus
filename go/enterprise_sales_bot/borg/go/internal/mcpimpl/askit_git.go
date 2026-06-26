package mcpimpl

import "context"

func HandleAskGit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	question, _ :=getString(args, "question")
	if question == "" {
		return err("question is required")
}

	return success("You asked: " + question)
}