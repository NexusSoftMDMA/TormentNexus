package tools

import "context"

func HandleSharplensLookup(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return success("Sharplens lookup result for: " + query)
}