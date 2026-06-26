package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleX_generic_mcp_client_chat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	response, e := http.DefaultClient.Get("http://example.com/api?message=" + message)
	if e != nil {
		return err("failed to send message")
}

	defer response.Body.Close()

	var result map[string]interface{}
	e = json.NewDecoder(response.Body).Decode(&result)
	if e != nil {
		return err("failed to decode response")
}

	return success("message sent successfully")
}

func HandleY_generic_mcp_client_chat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	chatID, _ :=getString(args, "chat_id")
	if chatID == "" {
		return err("chat_id is required")
}

	return success("chat_id received: " + chatID)
}