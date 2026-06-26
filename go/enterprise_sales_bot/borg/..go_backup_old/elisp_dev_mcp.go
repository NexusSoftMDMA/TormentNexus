package tools

import (
    "context"
    "fmt"
)

func HandleEvalElisp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    code, _ :=getString(args, "code")
    result := fmt.Sprintf("Evaluated: %s", code)
    return ok(result)
}

func HandleElispVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("Elisp Dev Mcp v1.0")
}