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
	"os"
)

func HandleWassengerSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	message, _ :=getString(args, "message")
	if to == "" || message == "" {
		return err("to and message are required")
}

	apiKey := os.Getenv("WASSENGER_API_KEY")
	if apiKey == "" {
		return err("WASSENGER_API_KEY environment variable not set")
}

	body := map[string]string{"to": to, "message": message


-reasoner (deepseek)*,
},
}