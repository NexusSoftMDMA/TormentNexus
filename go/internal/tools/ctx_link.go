//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandlePushContextBundle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	content, _ :=getString(args, "content")
	if name == "" || content == "" {
		return err("name and content are required")
}

	body, _ := json.Marshal(map[string]string{"name": name, "content": content})
	req, _ := http.NewRequestWithContext(ctx, "")


-reasoner (deepseek)*,
}