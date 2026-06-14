package tools

import (
	"context"
	"fmt"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	title, _ :=getString(args, "title")
	return ok(fmt.Sprintf("Pinned %s with title '%s'", url, title))
}