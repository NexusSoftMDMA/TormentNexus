package tools

import (
	"context"
)

func HandleKonceptGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	concept, _ :=getString(args, "concept")
	return success("got concept: " + concept)
}

func HandleKonceptSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return success("searching for: " + query)
}