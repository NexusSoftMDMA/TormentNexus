package tools

import (
	"context"
)

func HandleListDocuments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	docs := []string{"doc1.md", "doc2.txt", "report.pdf"}
	msg := "Documents: " + joinStrings(docs, ", ")
	return ok(msg)
}

func HandleSearchDocuments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if len(query) < 2 {
		return err("query must be at least 2 characters")
}

	docs := []string{"doc1.md", "doc2.txt", "report.pdf"}
	var matched []string
	for _, d := range docs {
		if contains(d, query) {
			matched = append(matched, d)

	}
	if len(matched) == 0 {
		return err("no documents found")
}

	msg := "Found: " + joinStrings(matched, ", ")
	return ok(msg)
}

}

func joinStrings(elems []string, sep string) string {
	if len(elems) == 0 {
		return ""
	}
	result := elems[0]
	for i := 1; i < len(elems); i++ {
		result += sep + elems[i]
	}
	return result
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}