package tools

import "context"

func HandleComponentSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok("Component search result for: " + query)
}

func HandleTokenLookup(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	return ok("Token value for: " + token)
}