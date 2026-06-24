package tools

import (
	"context"
	"errors"
	"math/rand"
	"strings"
)

func HandleFreeWill(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	choice := rand.Intn(2) == 0
	if choice {
		return success("Yes, you have free will")
}

	return success("No, it's predetermined")
}

func HandleMakeDecision(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	opts, _ :=getString(args, "options")
	if opts == "" {
		return err("options must be a non-empty comma-separated list")
}

	parts := strings.Split(opts, ",")
	if len(parts) == 0 {
		return err("no options provided")
}

	index := rand.Intn(len(parts))
	decision := strings.TrimSpace(parts[index])
	return success("I choose: " + decision)
}