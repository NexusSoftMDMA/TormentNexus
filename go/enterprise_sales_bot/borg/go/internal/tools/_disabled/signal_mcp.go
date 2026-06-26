package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	recipient, _ :=getString(args, "recipient")
	if recipient == "" {
		return err("recipient is required")
}

	baseURL := os.Getenv("SIGNAL_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	url := baseURL + "/v2/send"
	body := map[string]interface{}{
		"message":   message,
		"recipient": []string{recipient},
	}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewReader(jsonBody))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return err("bad status: " + fmt.Sprint(resp.StatusCode) + " " + string(respBody))
}

	return ok("message sent")
}