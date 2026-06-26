package tools

import (
	"context"
)

func HandleGoodNews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("The world is full of good news today!")
}