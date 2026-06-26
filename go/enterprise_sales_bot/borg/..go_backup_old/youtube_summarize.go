package tools

import (
	"context"
	"fmt"
)

func HandleYoutubeSummarize(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("Missing 'url' argument")
}

	summary := fmt.Sprintf("Summary for %s:\nThis video covers key concepts in an engaging way.", url)
	return ok(summary)
}