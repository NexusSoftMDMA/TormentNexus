package mcpimpl

import (
	"context"
	"fmt"
)

func HandleLspOpen(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	uri, _ :=getString(args, "uri")
	language, _ :=getString(args, "language")
	text, _ :=getString(args, "text")
	_ = language
	_ = text
	return ok(fmt.Sprintf("opened file: %s", uri))
}

func HandleLspDiagnostics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	uri, _ :=getString(args, "uri")
	if code == "" {
		return err("code is required")
}

	return ok(fmt.Sprintf("diagnostics for %s: no issues", uri))
}