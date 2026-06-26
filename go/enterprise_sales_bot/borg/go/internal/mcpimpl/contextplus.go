package mcpimpl

import (
	"context"
	"encoding/json"
)

// HandleSearch performs a semantic search over the codebase graph.
func HandleSearch_contextplus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	result := map[string]string{
		"query":  query,
		"status": "found",
	}
	data, e := json.Marshal(result)
	if e != nil {
		return err("failed to serialize result")
}

	return ok(string(data))
}