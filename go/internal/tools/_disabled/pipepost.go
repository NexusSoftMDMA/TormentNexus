package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	if to == "" {
		return err("missing 'to' field")
}

	message, _ :=getString(args, "message")
	if message == "" {
		return err("missing 'message' field")
}

	payload, e := json.Marshal(map[string]string{"to": to, "message": message})
	if e != nil {
		return err("failed to marshal payload: " + e.Error())
}

	resp, e := http.DefaultClient.Post("https://pipepost.example.com/api/send", "application/json", bytes.NewReader(payload))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return success("message sent to " + to)
}