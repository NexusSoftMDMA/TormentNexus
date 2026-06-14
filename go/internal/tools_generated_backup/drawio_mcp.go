package tools

import (
	"context"
	"encoding/base64"
	"fmt"
)

func HandleOpenDiagram(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	if content == "" {
		return err("content is required")
}

	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	url := fmt.Sprintf("https://app.diagrams.net/#H%s", encoded)
	return ok(fmt.Sprintf("Open this URL in your browser: %s", url))
}