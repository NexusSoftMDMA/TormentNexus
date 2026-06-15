//go:build ignore
// +build ignore

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
	token, _ :=getString(args, "access_token")
	chatID, _ :=getString(args, "chat_id")
	text, _ :=getString(args, "text")
	if token == "" || chatID == "" || text == "" {
		return err("Missing required arguments: access_token, chat_id, text")
	}
	body, _ := json.Marshal(map[string]interface{}{
		"receive_id": chatID,
		"msg_type":   "text",
		"content":    fmt.Sprintf(`{"text":"%s"}`,


-reasoner (deepseek)*,
},
}