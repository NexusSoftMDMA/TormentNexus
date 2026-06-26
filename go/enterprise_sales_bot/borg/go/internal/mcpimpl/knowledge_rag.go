package mcpimpl

import (
	"context"
)

// HandleSearch handles a knowledge RAG search query
func HandleSearch_knowledge_rag(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	maxResults, _ :=getInt(args, "maxResults")
	if maxResults <= 0 {
		maxResults = 5
	}
	result := "Knowledge RAG results for query: " + query + " (max results: " + string(rune(maxResults)) + ")"
	return ok(result)
}