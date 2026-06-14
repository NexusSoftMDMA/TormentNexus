package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetChatHistory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	chatID, _ :=getString(args, "chat_id")
	if chatID == "" {
		return err("chat_id is required")
}

	url := "https://chatlab.example.com/api/chat/" + chatID
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}

func HandleListChats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://chatlab.example.com/api/chats"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(fmt.Sprintf("Chats: %s", string(body)))
}