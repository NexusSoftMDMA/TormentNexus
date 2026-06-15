package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ :=getString(args, "user_id")
	text, _ :=getString(args, "text")
	if userID == "" || text == "" {
		return err("user_id and text are required")
}

	body, _ := json.Marshal(map[string]string{"user_id": userID, "text": text})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.yunxin.example.com/message/send", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to send message: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	return ok("message sent successfully")
}