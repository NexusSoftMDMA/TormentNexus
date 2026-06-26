package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListModels_mcp_server_ollama_bridge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://localhost:11434/api/tags")
	if e != nil {
		return err(fmt.Sprintf("failed to fetch models: %v", e))
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
}

	return ok(fmt.Sprintf("Models: %v", result["models"]))
}

func HandleGenerate_mcp_server_ollama_bridge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	prompt, _ :=getString(args, "prompt")
	reqBody, _ := json.Marshal(map[string]string{"model": model, "prompt": prompt})
	resp, e := http.DefaultClient.Post("http://localhost:11434/api/generate", "application/json", bytes.NewReader(reqBody))
	if e != nil {
		return err(fmt.Sprintf("generate request failed: %v", e))
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse error: %v", e))
}

	responseText, found := result["response"].(string)
	if !found {
		return err("missing response field")
}

	return ok(responseText)
}