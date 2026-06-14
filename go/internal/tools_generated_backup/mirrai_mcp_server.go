package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
)

func HandleGeneratePost(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	platform, _ :=getString(args, "platform")
	base := os.Getenv("MIRRA_API_BASE")
	if base == "" {
		base = "https://api.mirra.com"
	}
	body := map[string]interface{}{
		"prompt":   prompt,
		"platform": platform,
	}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, "POST", base+"/api/generate", bytes.NewReader(jsonBody))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("API request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("API returned non-200 status")
}

	var result struct {
		Content string `json:"content"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	return ok(result.Content)
}