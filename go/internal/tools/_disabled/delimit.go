package tools

import (
	"context"
	"strings"
)

func HandleSplit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	delimiter, _ :=getString(args, "delimiter")
	if delimiter == "" {
		delimiter = ","
	}
	parts := strings.Split(text, delimiter)
	return success(strings.Join(parts, "\n"))
}

func HandleJoin(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lines, _ :=getString(args, "lines")
	delimiter, _ :=getString(args, "delimiter")
	if delimiter == "" {
		delimiter = ","
	}
	parts := strings.Split(lines, "\n")
	return success(strings.Join(parts, delimiter))
}