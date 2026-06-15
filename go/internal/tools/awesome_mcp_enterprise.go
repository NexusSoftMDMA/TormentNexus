//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"strings"
)

func HandleList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	items := []map[string]string{
		{"name": "MCP CLI", "description": "Command-line interface for MCP"},
		{"name": "MCP Hub", "description": "Centralized tool registry"},
	}
	data, e := json.Marshal(items)
	if e != nil {
		return err("failed to marshal list")
}

	return success(string(data))
}

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	items := []map[string]string{
		{"name": "MCP CLI", "description": "Command-line interface for MCP"},
		{"name": "MCP Hub", "description": "Centralized tool registry"},
	}
	var filtered []map[string]string
	for _, item := range items {
		if strings.Contains(item["name"], query) || strings.Contains(item["description"], query) {
			filtered = append(filtered, item)

	}
	data, e := json.Marshal(filtered)
	if e != nil {
		return err("failed to marshal search results")
}

	return success(string(data))
}
}
