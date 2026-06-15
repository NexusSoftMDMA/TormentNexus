package tools

import (
	"context"
	"net/http"
)

func HandleParseCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = http.DefaultClient
	code, _ :=getString(args, "code")
	return success("Parsed code: " + code)
}

func HandleGetSymbols(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = http.DefaultClient
	code, _ :=getString(args, "code")
	return success("Symbols: " + code)
}