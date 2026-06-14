package tools

import "context"

func HandleMnemoAddFact(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fact, _ :=getString(args, "fact")
	if fact == "" {
		return err("fact is required")
}

	return ok("fact recorded: " + fact)
}

func HandleMnemoQueryFact(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	return success("query result placeholder for: " + query)
}// touch 1781132135
