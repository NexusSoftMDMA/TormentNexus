package tools

import (
	"context"
)

func HandleHumanize(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
	}
	humanized := "Dear User, you said: " + text
	return ok(humanized)
}

func HandleClarify(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	question, _ :=getString(args, "question")
	if question == "" {
		return err("question is required")
	}
	clarification := "I understand your question: '" + question + "'. Let me explain..."
	return ok(clarification)
}// touch 1781132138
