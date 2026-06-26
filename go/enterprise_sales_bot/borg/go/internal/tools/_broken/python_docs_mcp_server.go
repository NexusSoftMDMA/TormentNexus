package mcpimpl

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// HandlePythonDocs returns documentation for a given Python topic.
func HandlePythonDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic := strings.ToLower(getString(args, "topic"))
	docs := map[string]string{
		"list":   "A mutable sequence type.",
		"dict":   "A key-value mapping type.",
		"set":    "An unordered collection of unique elements.",
		"tuple":  "An immutable sequence type.",
		"string": "An immutable sequence of characters.",
	}
	if doc, found := docs[topic]; found {
		return ok(doc)
}

	return err("Topic not found")
}

// HandlePythonSearch returns a search URL for Python documentation.
func HandlePythonSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("Query is required")
}

	url := fmt.Sprintf("https://docs.python.org/3/search.html?q=%s", query)
	return ok(url)
}