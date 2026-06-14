package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleChat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	body := map[string]string{"prompt": prompt}
	data, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request")
}

	resp, e := http.DefaultClient.Post("https://api.anythingllm.com/chat", "application/json", bytes.NewReader(data))
	if e != nil {
		return err("failed to call API: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Response string `json:"response"`
	}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	return ok(result.Response)
}