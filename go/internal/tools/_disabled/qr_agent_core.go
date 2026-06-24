package tools

import "context"

func HandleGenerateQR(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	return success("QR code generated for: " + text)
}