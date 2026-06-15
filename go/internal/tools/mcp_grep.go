//go:build ignore
// +build ignore

package tools

import (
	"context"
	"strings"
)

func HandleGrep(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pattern, _ :=getString(args, "pattern")
	content, _ :=getString(args, "content")
	ignoreCase, _ :=getBool(args, "ignoreCase")

	if pattern == "" {
		return err("pattern is required")
}

	if content == "" {
		return err("content is required")
}

	lines := strings.Split(content, "\n")
	var matches []string

	for _, line := range lines {
		match := false
		if ignoreCase {
			match = strings.Contains(strings.ToLower(line), strings.ToLower(pattern))
		} else {
			match = strings.Contains(line, pattern)

		if match {
			matches = append(matches, line)

	}

	return ok(strings.Join(matches, "\n"))
}
}
}
