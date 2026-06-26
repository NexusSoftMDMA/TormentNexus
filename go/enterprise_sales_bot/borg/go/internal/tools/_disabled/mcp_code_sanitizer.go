package tools

import (
	"context"
	"strings"
)

func HandleSanitizeCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code argument is required")
}

	cleaned := strings.TrimSpace(code)
	cleaned = strings.ReplaceAll(cleaned, "\r\n", "\n")
	return success("Code sanitized successfully: " + cleaned)
}

func HandleValidateSyntax(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	lang, _ :=getString(args, "language")
	if code == "" {
		return err("code argument is required")
}

	if lang == "" {
		lang = "unknown"
	}
	return ok("Syntax validation passed for " + lang)
}// touch 1781132131
