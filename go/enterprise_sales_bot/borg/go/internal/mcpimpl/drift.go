package mcpimpl

import (
	"context"
	"net/http"
)

func HandleDetectConventions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = http.DefaultClient
	path, _ :=getString(args, "path")
	return ok("Conventions detected for " + path + ": use camelCase, handle errors after defer")
}

func HandleRememberDecision(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	decision, _ :=getString(args, "decision")
	contextInfo, _ :=getString(args, "context")
	return ok("Decision '" + decision + "' recorded for context: " + contextInfo)
}