package tools

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func HandleGetTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	format, _ :=getString(args, "format")
	if format == "" {
		format = time.RFC3339
	}
	now := time.Now().Format(format)
	return ok(fmt.Sprintf("Current time: %s", now))
}

func HandleSum(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	return ok(fmt.Sprintf("Sum: %d", a+b))
}