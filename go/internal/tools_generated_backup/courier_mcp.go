package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleSendNotification(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	recipient, _ :=getString(args, "recipient")
	message, _ :=getString(args, "message")
	if recipient == "" || message == "" {
		return err("recipient and message are required")
}

	body := map[string]string{"recipient": recipient, "message": message}
	data, _ := json.Marshal(body)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.courier.com/send", strings.NewReader(string(data)))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	return ok("notification sent")
}