package tools

import (
	"context"
)

func HandleSearchClasses(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = ctx
	query, _ :=getString(args, "query")
	msg := "Common TailwindCSS utility classes: flex, m-4, bg-blue-500, text-center, p-4, etc."
	if query != "" {
		msg = "Classes matching '" + query + "': flex, m-4 (example)"
	}
	return ok(msg)
}

func HandleGetClassDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = ctx
	className, _ :=getString(args, "class")
	if className == "" {
		return err("Missing required argument 'class'")
}

	docs := "Documentation for class '" + className + "': This utility applies ... (static example)"
	return ok(docs)
}