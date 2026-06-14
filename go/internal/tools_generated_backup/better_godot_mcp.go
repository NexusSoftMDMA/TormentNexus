package tools

import (
	"context"
	"fmt"
)

func HandleGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return success("Hello, " + name + "!")
}

func HandleAdd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	sum := a + b
	return ok(fmt.Sprintf("Sum: %d", sum))
}