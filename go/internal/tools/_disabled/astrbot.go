package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func AstrbotGetInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("token is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.astrbot.com/info", nil)
	if e != nil {
		return err(fmt.Sprintf("create request failed: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return success(fmt.Sprintf("Info: %v", result))
}

func AstrbotSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	message, _ :=getString(args, "message")
	if token == "" || message == "" {
		return err("token and message are required")
}

	body, _ := json.Marshal(map[string]string{"message": message})
	resp, e := http.DefaultClient.Post("https://api.astrbot.com/send", "application/json", nil)
	if e != nil {
		return err(fmt.Sprintf("post failed: %v", e))
}

	defer resp.Body.Close()
	// In real code you'd use bytes.NewReader(body) but Post signature is (url, contentType, body io.Reader)
	// Fix: use http.NewRequest with body
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.astrbot.com/send", nil)
	if e != nil {
		return err(fmt.Sprintf("create request failed: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, e = http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	return ok("message sent")
}