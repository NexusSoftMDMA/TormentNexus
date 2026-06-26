package mcpimpl

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func HandleGetCurrentTime_five_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().Format(time.RFC3339)
	return success(fmt.Sprintf("Current time: %s", now))
}

func HandleGetRandomNumber(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(100) + 1
	return success(fmt.Sprintf("Random number: %d", num))
}