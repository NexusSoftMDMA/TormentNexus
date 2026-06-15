package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGenerate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	model, _ :=getString(args, "model")
	if model == "" {
		model = "llama2"
	}
	payload := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:11434/api/generate", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http error: %v", e))
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read error: %v", e))
}

	var result map[string]interface{}
	e = json.Unmarshal(respBody, &result)
	if e != nil {
		return err(fmt.Sprintf("parse error: %v", e))
}

	responseText, found := result["response"].(string)
	if !found {
		return err("no response field in Ollama output")
}

	return ok(responseText)
}