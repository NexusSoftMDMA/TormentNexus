package mcpimpl

import (
	"context"
	"math/rand"
	"time"
)

func HandleDecide(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	question, _ :=getString(args, "question")
	if question == "" {
		question = "Should I?"
	}
	rand.Seed(time.Now().UnixNano())
	answers := []string{"Yes", "No", "Maybe", "Ask again later", "Definitely", "Absolutely not"}
	choice := answers[rand.Intn(len(answers))]
	return success(question + " → " + choice)
}