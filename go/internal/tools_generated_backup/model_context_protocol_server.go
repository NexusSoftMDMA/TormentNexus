package tools

import (
	"context"
	"fmt"
)

func HandleGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok(fmt.Sprintf("Hello, %s!", name))
}

func HandleAdd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	return ok(fmt.Sprintf("%d + %d = %d", a, b, a+b))
}