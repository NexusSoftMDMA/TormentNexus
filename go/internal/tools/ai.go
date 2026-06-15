//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleAsk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	resp, e := http.DefaultClient.Get("https://ai.example.com/ask?q=" + url.QueryEscape(prompt))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return ok(string(body))
}

	answer, found := result["answer"]
	if !found {
		return ok("no answer")
}

	return ok(fmt.Sprintf("%v", answer))
}

func HandleGenerate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	resp, e := http.DefaultClient.PostForm("https://ai.example.com/generate", url.Values{"prompt": {prompt}})
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return ok(string(body))
}

	text, found := result["text"]
	if !found {
		return ok("no text")
}

	return ok(fmt.Sprintf("%v", text))
}
