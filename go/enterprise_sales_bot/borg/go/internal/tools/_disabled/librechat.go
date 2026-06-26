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

	model, _ :=getString(args, "model")
	body := map[string]interface{}{"message": message}
	if model != "" {
		body["model"] = model
	}
	jsonBytes, e := json.Marshal(body)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:3000/api/chat/send", bytes.NewReader(jsonBytes))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return success(string(respBody))
}