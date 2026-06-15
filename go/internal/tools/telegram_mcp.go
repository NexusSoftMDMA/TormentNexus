//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func HandleSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	chatID, _ :=getString(args, "chat_id")
	text, _ :=getString(args, "text")
	if chatID == "" || text == "" {
		return err("chat_id and text required")
}


-reasoner (deepseek)*
}