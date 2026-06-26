package tools

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func HandleGetCurrentTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().Format(time.RFC3339)
	return ok(fmt.Sprintf("Current time is %s", now))
}

func HandleGenerateRandomNumber(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	min, _ :=getInt(args, "min")
	max, _ :=getInt(args, "max")
	if max <= min {
		return err("max must be greater than min")
}

	num := rand.Intn(max-min) + min
	return ok(fmt.Sprintf("Random number: %d", num))
}