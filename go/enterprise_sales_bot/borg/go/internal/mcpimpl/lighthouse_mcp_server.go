package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleAskLighthouse(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	reqBody := map[string]string{"prompt": prompt}
	body, e := json.Marshal(reqBody)
	if e != nil {
		return err("failed to marshal request")
}

	resp, e := http.DefaultClient.Post("https://api.lighthouse.ai/v1/chat/completions", "application/json", strings.NewReader(string(body)))
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	e = json.Unmarshal(respBody, &result)
	if e != nil {
		return err("failed to parse response")
}

	choices, found := result["choices"].([]interface{})
	if !found || len(choices) == 0 {
		return err("no choices in response")
}

	choice, found := choices[0].(map[string]interface{})
	if !found {
		return err("invalid choice format")
}

	text, found := choice["text"].(string)
	if !found {
		return err("no text in choice")
}

	return ok(text)
}

func HandleGetModels_lighthouse_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.lighthouse.ai/v1/models")
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var models []string
	e = json.Unmarshal(body, &models)
	if e != nil {
		return err("failed to parse models")
}

	return ok(fmt.Sprintf("Models: %v", models))
}