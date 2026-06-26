package tools

import "context"

func HandleFormatCitation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	style, _ :=getString(args, "style")
	data, _ :=getString(args, "data")
	if style == "" || data == "" {
		return err("style and data are required")
}

	return ok("Formatted citation: " + style + " - " + data)
}