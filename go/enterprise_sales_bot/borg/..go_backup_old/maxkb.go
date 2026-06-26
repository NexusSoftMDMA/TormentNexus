package tools

import (
	"context"
)

func HandleMaxKBQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		return err("query is required")
}

	return success("processed query: " + q)
}

func HandleMaxKBList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		return err("limit must be positive")
}

	return ok("list with limit " + string(rune(limit))) // note: simplistic conversion
}