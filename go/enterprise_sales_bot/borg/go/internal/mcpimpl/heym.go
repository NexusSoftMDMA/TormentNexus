package mcpimpl

import (
	"context"
	"fmt"
)

func HandleEcho_heym(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(fmt.Sprintf("Echo: %s", msg))
}

func HandleReverse_heym(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	runes := []rune(msg)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return ok(fmt.Sprintf("Reversed: %s", string(runes)))
}