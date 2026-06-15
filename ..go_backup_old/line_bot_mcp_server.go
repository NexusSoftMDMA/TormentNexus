package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleReplyMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "replyToken")
	text, _ :=getString(args, "message")
	body, e := json.Marshal(map[string]interface{}{
		"replyToken": token,
		"messages":   []map[string]string{{"type": "text", "text": text}},
	})
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.line.me/v2/bot/message/reply", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+getString(args, "channelAccessToken"))
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("api error: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("LINE API error: %s", string(b)))
}

	return ok("Message replied")
}

func HandlePushMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	text, _ :=getString(args, "message")
	body, e := json.Marshal(map[string]interface{}{
		"to":       to,
		"messages": []map[string]string{{"type": "text", "text": text}},
	})
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.line.me/v2/bot/message/push", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+getString(args, "channelAccessToken"))
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("api error: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("LINE API error: %s", string(b)))
}

	return ok("Message pushed")
}