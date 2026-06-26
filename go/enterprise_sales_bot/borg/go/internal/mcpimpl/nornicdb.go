package mcpimpl

import (
	"context"
	"encoding/json"
)

func HandleExecuteCypher_nornicdb(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	result := map[string]interface{}{
		"query":  query,
		"status": "executed",
		"rows":   0,
	}
	data, e := json.Marshal(result)
	if e != nil {
		return err("failed to marshal result")
}

	return ok(string(data))
}

func HandleVectorSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	collection, _ :=getString(args, "collection")
	vecStr, _ :=getString(args, "vector")
	limit, _ :=getInt(args, "limit")
	if collection == "" || vecStr == "" {
		return err("collection and vector are required")
}

	result := map[string]interface{}{
		"collection": collection,
		"vector":     vecStr,
		"limit":      limit,
		"status":     "searched",
		"matches":    0,
	}
	data, e := json.Marshal(result)
	if e != nil {
		return err("failed to marshal result")
}

	return ok(string(data))
}