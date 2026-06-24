package tools

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleGenerateQrCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	qrURL := "https://api.qrserver.com/v1/create-qr-code/?size=150x150&data=" + url.QueryEscape(text)
	resp, e := http.DefaultClient.Get(qrURL)
	if e != nil {
		return err("failed to call QR API: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok("QR code generated. Image data: " + string(body))
}