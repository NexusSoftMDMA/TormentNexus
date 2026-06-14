package tools

import (
	"context"
	"fmt"
	"strings"
)

func HandleModularRag(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	docs := []string{
		"The sky is blue and vast",
		"Roses are red, violets are blue",
		"Grass is green in spring",
		"Modular systems are flexible",
		"RAG combines retrieval and generation",
	}
	var results []string
	for _, d := range docs {
		if strings.Contains(strings.ToLower(d), strings.ToLower(query)) {
			results = append(results, d)

	}
	if len(results) == 0 {
		return success(fmt.Sprintf("No documents found for query: %s", query))
}

	return success(fmt.Sprintf("Found %d document(s): %s", len(results), strings.Join(results, "; ")))
}
}