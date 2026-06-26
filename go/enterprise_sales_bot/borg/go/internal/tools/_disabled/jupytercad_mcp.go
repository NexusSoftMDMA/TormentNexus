package tools

import (
	"context"
	"fmt"
)

func HandleExecuteCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	return ok(fmt.Sprintf("Executing code: %s", code))
}

func HandleGetNotebooks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok(`["notebook1.ipynb", "notebook2.ipynb"]`)
}