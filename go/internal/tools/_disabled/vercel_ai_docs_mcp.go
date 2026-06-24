package tools

import "context"

func HandleSearchVercelDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	return success("Search results for: " + query)
}

func HandleGetVercelDoc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	if topic == "" {
		return err("topic parameter is required")
}

	return ok("Documentation for " + topic + " is available at https://vercel.com/docs/ai")
}