//go:build ignore
// +build ignore

package tools

import (
	"context"
	"strconv"
)

func HandleGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok("Hello, " + name + "!")
}

func HandleLength(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	length := strconv.Itoa(len(text))
	return ok("Length: " + length)
}
