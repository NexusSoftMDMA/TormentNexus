package tools

import (
	"context"
	"fmt"
)

func HandleSayHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return ok(fmt.Sprintf("Hello, %s!", name))
}