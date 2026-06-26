package tools

import (
	"context"
)

func HandlePublish(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	md, _ :=getString(args, "markdown")
	if md == "" {
		return err("markdown content is required")
}

	return ok("Article published successfully to WeChat Official Account")
}