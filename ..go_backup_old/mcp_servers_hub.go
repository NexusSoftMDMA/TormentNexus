package tools

import (
	"context"
)

func ListMcpServersHandler(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filter, _ :=getString(args, "filter")
	servers := []string{
		"server-a",
		"server-b",
		"server-c",
	}
	if filter != "" {
		var matched []string
		for _, s := range servers {
			if s == filter {
				matched = append(matched, s)

		}
		if len(matched) == 0 {
			return ok("No matching MCP servers found")
}

		return ok("Matching MCP servers: " + joinStrings(matched))
}

	return ok("Available MCP servers: " + joinStrings(servers))
}

}

func joinStrings(items []string) string {
	result := ""
	for i, s := range items {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}