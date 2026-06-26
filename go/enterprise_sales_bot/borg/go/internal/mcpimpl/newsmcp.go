package mcpimpl

import (
	"context"
	"fmt"
)

func HandleGetTopNews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")
	news := fmt.Sprintf("Top news in category '%s': No API configured, returning mock data", category)
	return ok(news)
}

func HandleSearchNews_newsmcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	result := fmt.Sprintf("Search results for '%s': No API configured, returning mock data", query)
	return success(result)
}