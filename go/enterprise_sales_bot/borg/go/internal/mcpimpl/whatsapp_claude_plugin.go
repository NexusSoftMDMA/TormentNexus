package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSendWhatsAppMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	message, _ :=getString(args, "message")
	if to == "" || message == "" {
		return err("to and message are required")
}

	apiURL := fmt.Sprintf("https://api.whatsapp.com/send?phone=%s&text=%s", url.QueryEscape(to), url.QueryEscape(message))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err(fmt.Sprintf("failed to send message: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(body)))
}

	return ok("Message sent successfully")
}