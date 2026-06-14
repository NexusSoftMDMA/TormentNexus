package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	message, _ :=getString(args, "message")
	if to == "" || message == "" {
		return err("missing required arguments: to, message")
}

	body := map[string]string{"to": to, "message": message}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal request: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.ycloud.com/v1/whatsapp/messages", bytes.NewReader(jsonBody))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	return success("message sent successfully")
}