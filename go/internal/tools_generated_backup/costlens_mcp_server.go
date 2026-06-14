package tools

import (
	"context"
	"fmt"
	"strings"
)

func HandleClassifyPrompt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	words := strings.Fields(prompt)
	wordCount := len(words)
	charCount := len(prompt)
	var complexity string
	var costFactor float64
	if wordCount < 10 && charCount < 100 {
		complexity = "low"
		costFactor = 0.5
	} else if wordCount < 50 && charCount < 500 {
		complexity = "medium"
		costFactor = 1.0
	} else {
		complexity = "high"
		costFactor = 2.0
	}
	msg := fmt.Sprintf("Complexity: %s, estimated cost factor: %.1f", complexity, costFactor)
	return ok(msg)
}