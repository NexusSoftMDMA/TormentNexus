package tools

import (
	"context"
	"net/http"
	"encoding/json"
)

func HandleGetFact(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	animal, _ :=getString(args, "animal")
	if animal == "" {
		animal = "peacock"
	}
	fact := animal + "s are known for their colorful feathers."
	return ok(fact)
}

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}