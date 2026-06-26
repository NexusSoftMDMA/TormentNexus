package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
)

func HandleListConversations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := os.Getenv("MESSAGES_API_BASE")
	if base == "" {
		base = "http://localhost:8080"
	}
	resp, e := http.DefaultClient.Get(base + "/conversations")
	if e != nil {
		return err("failed to fetch conversations: " + e.Error())
}

	defer resp.Body.Close()
	var data []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok("conversations retrieved successfully")
}

func HandleSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	recipient, _ :=getString(args, "recipient")
	text, _ :=getString(args, "text")
	if recipient == "" || text == "" {
		return err("recipient and text are required")
}

	base := os.Getenv("MESSAGES_API_BASE")
	if base == "" {
		base = "http://localhost:8080"
	}
	body, _ := json.Marshal(map[string]string{"recipient": recipient, "text": text})
	resp, e := http.DefaultClient.Post(base+"/send", "application/json", jsonReader(body))
	if e != nil {
		return err("failed to send message: " + e.Error())
}

	resp.Body.Close()
	return success("message sent successfully")
}

func jsonReader(data []byte) *bytes.Buffer {
	return bytes.NewBuffer(data)
}