package tools

import (
	"context"
	"fmt"
)

func HandleHey(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return success(fmt.Sprintf("Hey, %s!", name))
}