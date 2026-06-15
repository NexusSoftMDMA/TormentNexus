package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleConcilium(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	msg := fmt.Sprintf("Claude Concilium says hello to %s", name)
	return ok(msg)
}