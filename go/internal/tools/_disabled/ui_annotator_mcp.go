package tools

import (
	"context"
)

func HandleAnnotate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	msg := "Annotated URL: " + url
	return ok(msg)
}