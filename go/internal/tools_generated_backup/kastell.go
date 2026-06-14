package tools

import (
	"context"
	"fmt"
)

func HandleKastellInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return success(fmt.Sprintf("Hello %s from Kastell!", name))
}