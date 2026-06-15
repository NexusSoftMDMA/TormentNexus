package tools

import (
	"context"
	"math/rand"
	"time"
)

func HandleGetFact(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	facts := []string{
		"The first computer virus was created in 1983.",
		"A day on Venus is longer than a year on Venus.",
		"Octopuses have three hearts.",
	}
	idx := rand.Intn(len(facts))
	return ok(facts[idx])
}

func HandleGetTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	format, _ :=getString(args, "format")
	if format == "" {
		format = time.RFC3339
	}
	now := time.Now().Format(format)
	return ok(now)
}