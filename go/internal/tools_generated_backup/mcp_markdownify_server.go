package tools

import (
	"context"
	"regexp"
	"strings"
)

func HandleConvertToMarkdown(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	html, _ :=getString(args, "html")
	if html == "" {
		return err("html argument is required")
}

	// Replace <br> with newline
	reBr := regexp.MustCompile(`(?i)<br\s*/?>`)
	html = reBr.ReplaceAllString(html, "\n")
	// Remove all remaining HTML tags
	reTag := regexp.MustCompile(`<[^>]*>`)
	plain := reTag.ReplaceAllString(html, "")
	plain = strings.TrimSpace(plain)
	return success(plain)
}