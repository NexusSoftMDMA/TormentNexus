package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.genai.example.com/models")
	if e != nil {
		return err("failed to fetch models: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result struct {
		Models []string `json:"models"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse: " + e.Error())
}

	return ok(fmt.Sprintf("Available models: %v", result.Models))
}

func HandleGenerateText(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	model, _ :=getString(args, "model")
	if prompt == "" {
		return err("prompt is required")
}

	payload := map[string]string{"prompt": prompt, "model": model}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	resp, e := http.DefaultClient.Post("https://api.genai.example.com/generate", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	return ok(fmt.Sprintf("Response: %s", string(respBody)))
}