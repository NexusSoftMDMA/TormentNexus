package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleDeepseekChat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	messages := args["messages"].([]interface{})
	reqBody := map[string]interface{}{
		"model":    "deepseek-chat",
		"messages": messages,
	}
	body, e := json.Marshal(reqBody)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	httpReq, e := http.NewRequestWithContext(ctx, "POST", "https://api.deepseek.com/v1/chat/completions", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	httpReq.Header.Set("Content-Type", "application/json")
	httpResp, e := http.DefaultClient.Do(httpReq)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer httpResp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(httpResp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("%v", result))
}

func HandleDeepseekReasoner(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	messages := args["messages"].([]interface{})
	reqBody := map[string]interface{}{
		"model":    "deepseek-reasoner",
		"messages": messages,
	}
	body, e := json.Marshal(reqBody)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	httpReq, e := http.NewRequestWithContext(ctx, "POST", "https://api.deepseek.com/v1/chat/completions", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	httpReq.Header.Set("Content-Type", "application/json")
	httpResp, e := http.DefaultClient.Do(httpReq)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer httpResp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(httpResp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("%v", result))
}