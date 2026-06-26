package tools

import "context"

func HandleSearchDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	results := "Search results for: " + query
	return success(results)
}

func HandleGetDoc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	docID, _ :=getString(args, "docId")
	if docID == "" {
		return err("docId is required")
}

	content := "Document content for: " + docID
	return ok(content)
}