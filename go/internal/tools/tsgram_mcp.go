//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetMe(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("token required")
}

	url := "https://api.telegram.org/bot")


-reasoner (deepseek)*
}