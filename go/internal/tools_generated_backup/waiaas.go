package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGenerateText(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.waiaas.com/generate?prompt="+prompt, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to call API")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON")
}

	text, found := result["text"].(string)
	if !found {
		return err("missing text field")
}

	return ok(fmt.Sprintf("Generated text: %s", text))
}

func HandleGetModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.waiaas.com/models", nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to call API")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var models []string
	if e := json.Unmarshal(body, &models); e != nil {
		return err("invalid JSON")
}

	return ok(fmt.Sprintf("Available models: %v", models))
}