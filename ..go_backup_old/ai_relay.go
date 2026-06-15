package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleChatCompletions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	model, _ :=getString(args, "model")
	messages := args["messages"]
	b, e := json.Marshal(map[string]interface{}{
		"model":    model,
		"messages": messages,
	})
	if e != nil {
		return err("marshal error: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", io.NopCloser(io.LimitReader(nil, 0)))
	if e != nil {
		return err("request error: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API error: " + string(body))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("unmarshal error: " + e.Error())
}

	return ok(fmt.Sprintf("%v", result))
}

func HandleAnthropicMessages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	model, _ :=getString(args, "model")
	messages := args["messages"]
	bodyMap := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"max_tokens": getInt(args, "max_tokens"),
	}
	b, e := json.Marshal(bodyMap)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", io.NopCloser(nil))
	if e != nil {
		return err("request error: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API error: " + string(respBody))
}

	var result map[string]interface{}
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err("unmarshal error: " + e.Error())
}

	return ok(fmt.Sprintf("%v", result))
}