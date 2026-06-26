package mcpimpl

import "context"

func HandleAskQuestion_mk_qa_master(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	question, _ :=getString(args, "question")
	return ok("Answer: " + question + " (from Mk Qa Master)")
}