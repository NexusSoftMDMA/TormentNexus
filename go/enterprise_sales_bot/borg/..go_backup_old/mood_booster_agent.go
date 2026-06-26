package tools

import (
	"context"
	"math/rand"
	"time"
)

func HandleMood(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	quotes := []string{
		"You are awesome!",
		"Keep going!",
		"Smile, it's contagious!",
		"You've got this!",
		"Today is a great day!",
	}
	rand.Seed(time.Now().UnixNano())
	idx := rand.Intn(len(quotes))
	return success(quotes[idx])
}