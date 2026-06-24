package tools

import (
	"context"
	"strings"
)

func HandleSieve(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	logs, _ :=getString(args, "logs")
	pattern, _ :=getString(args, "pattern")
	if pattern == "" {
		return err("pattern is required")
}

	var filtered []string
	for _, line := range strings.Split(logs, "\n") {
		if strings.Contains(line, pattern) {
			filtered = append(filtered, line)

	}
	return success(strings.Join(filtered, "\n"))
}
}