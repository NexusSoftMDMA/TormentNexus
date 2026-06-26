package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleChat_mcp_client(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("missing prompt")
}

	body := map[string]interface{}{"contents": []map[string]interface{}{{"parts": []map[string]interface{}{{"text": prompt}}}}}
	payload, e := json.Marshal(body)
	if e != nil {
		return err("marshal failed")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=API_KEY", nil)
	if e != nil {
		return err("request failed")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("do failed")
}

	defer resp.Body.Close()
	return success("chat completed")
}

func HandleGraph(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	if input == "" {
		return err("missing input")
}

	return success("graph processed")
}