package mcpimpl

import (
	"context"
	"net/url"
)

func HandleGenerateQrCode_qrcode_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text parameter is required")
}

	qrURL := "https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=" + url.QueryEscape(text)
	return success("QR Code generated: " + qrURL)
}