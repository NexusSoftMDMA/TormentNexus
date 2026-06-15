package tools

import (
	"context"
	"fmt"
)

func HandleRekindle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	msg := fmt.Sprintf("Rekindled: %s - remember the spark!", topic)
	return ok(msg)
}

func HandleRekindleQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = getString(args, "type")
	return ok("Rekindle your passion with this quote: 'The best time to plant a tree was 20 years ago.'")
}