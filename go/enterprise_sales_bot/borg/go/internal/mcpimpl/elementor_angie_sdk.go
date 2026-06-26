package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleAngieChat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	apiKey, _ :=getString(args, "apiKey")

	body := map[string]string{"prompt": prompt}
	jsonData, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.angie.ai/v1/chat", bytes.NewBuffer(jsonData))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()

	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err("failed to parse response")
}

	reply, found := result["reply"].(string)
	if !found {
		return err("response missing reply field")
}

	return ok(reply)
}