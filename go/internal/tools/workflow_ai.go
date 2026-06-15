//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func HandleAI(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("missing prompt")
}

	reqBody := map[string]string{"prompt": prompt}
	bodyBytes, e


-reasoner (deepseek)*,
}