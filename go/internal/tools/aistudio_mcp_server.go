//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleGenerateContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	prompt, _ :=getString(args, "prompt")
	if apiKey == "" || prompt == "" {
		return err("apiKey and prompt are required")
}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=%s", apiKey)
	body := fmt.Sprintf(`{"contents":[{"parts":[{"text":"%s"}]}]}`, strings.ReplaceAll(prompt, `"`, `\"`))
	resp, e := http.DefaultClient.Post(url, "application/json", strings.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	jsonStr, _ := json.Marshal(result)
	return ok(string(jsonStr))
}

func HandleListModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey is required")
}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	jsonStr, _ := json.Marshal(result)
	return ok(string(jsonStr))
}