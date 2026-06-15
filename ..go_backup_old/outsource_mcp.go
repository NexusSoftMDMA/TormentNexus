package tools

import (
	"context"
	"net/http"
)

func HandleGreeting(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "user"
	}
	resp := map[string]interface{}{"message": "Hello " + name + ", welcome to Outsource Mcp!"}
	_ = http.DefaultClient // avoid unused import
	return ok(resp)
}