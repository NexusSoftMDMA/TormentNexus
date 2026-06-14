package tools

import "context"

func HandleRecall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok("recalled: " + query)
}

func HandleRemember(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fact, _ :=getString(args, "fact")
	return ok("memorized: " + fact)
}