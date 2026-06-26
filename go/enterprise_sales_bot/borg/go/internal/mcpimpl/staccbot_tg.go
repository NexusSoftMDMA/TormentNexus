package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSendTelegramMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	chatID, _ :=getString(args, "chat_id")
	text, _ :=getString(args, "text")
	if token == "" || chatID == "" || text == "" {
		return err("Missing required parameters")
}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	data := url.Values{}
	data.Set("chat_id", chatID)
	data.Set("text", text)
	resp, e := http.DefaultClient.PostForm(apiURL, data)
	if e != nil {
		return err("Failed to send message: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return err("Telegram API error: " + resp.Status + " - " + string(body))
}

	return ok("Message sent successfully")
}