package tools

import "context"

func HandleSearchLogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	return ok("search initiated for: " + query)
}

func HandleListLogTopics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	region, _ :=getString(args, "region")
	if region == "" {
		return err("region is required")
}

	return ok("listing topics in region: " + region)
}