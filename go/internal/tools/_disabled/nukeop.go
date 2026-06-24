package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleNuke(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	_ = http.DefaultClient
	msg := fmt.Sprintf("Nukeop says hello, %s!", name)
	return success(msg)
}