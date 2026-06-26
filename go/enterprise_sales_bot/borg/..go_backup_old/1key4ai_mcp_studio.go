package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	body, _ := json.Marshal(map[string]interface{}{"message": message})
	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/api/message", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, string(respBody)))
}

	return ok("message sent successfully")
}

func HandleManageFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	path, _ :=getString(args, "path")
	if action == "" || path == "" {
		return err("action and path are required")
}

	payload := map[string]interface{}{"action": action, "path": path}
	if content := getString(args, "content"); content != "" {
		payload["content"] = content
	}
	body, _ := json.Marshal(payload)
	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/api/file", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, string(respBody)))
}

	return ok("file operation completed")
}