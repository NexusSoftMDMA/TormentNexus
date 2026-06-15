//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
)

func HandleListResources(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")
	resources := []map[string]string{
		{"name": "MCP Specification", "url": "https://spec.modelcontextprotocol.io"},
		{"name": "Awesome MCP List", "url": "https://github.com/awesome-mcp"},
		{"name": "MCP Documentation", "url": "https://docs.modelcontextprotocol.io"},
	}
	if category != "" {
		filtered := []map[string]string{}
		for _, r := range resources {
			if r["name"] == category || r["url"] == category {
				filtered = append(filtered, r)

		}
		resources = filtered
	}
	return success(map[string]interface{}{"resources": resources})
}

}

func HandleSearchResources(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
	}
	results := []map[string]string{
		{"name": "MCP Specification", "url": "https://spec.modelcontextprotocol.io"},
		{"name": "Awesome MCP List", "url": "https://github.com/awesome-mcp"},
	}
	matched := []map[string]string{}
	for _, r := range results {
		if r["name"] == query || r["url"] == query {
			matched = append(matched, r)

	}
	return success(map[string]interface{}{"results": matched})
}
}
