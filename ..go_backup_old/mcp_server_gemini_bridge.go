package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGeminiPrompt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return err("GEMINI_API_KEY not set")
}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=" + apiKey
	body := map[string]interface{}{
		"contents": []interface{}{
			map[string]interface{}{
				"parts": []interface{}{
					map[string]interface{}{"text": prompt},
				},
			},
		},
	}
	bodyBytes, e := json.Marshal(body)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if e := json.Unmarshal(data, &result); e != nil {
		return err("unmarshal error: " + e.Error())
}

	candidates, found := result["candidates"].([]interface{})
	if !found || len(candidates) == 0 {
		return err("no candidates in response")
}

	parts, found := candidates[0].(map[string]interface{})["content"].(map[string]interface{})["parts"].([]interface{})
	if !found || len(parts) == 0 {
		return err("no parts in response")
}

	text, found := parts[0].(map[string]interface{})["text"].(string)
	if !found {
		return err("text not found in response")
}

	return ok(text)
}