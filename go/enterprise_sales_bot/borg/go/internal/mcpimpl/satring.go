package mcpimpl

import (
	"context"
	"strconv"
)

func HandleReverse_satring(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	s, _ :=getString(args, "text")
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return ok(string(runes))
}

func HandleLength(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	s, _ :=getString(args, "text")
	return ok(strconv.Itoa(len([]rune(s))))
}