package tools

import (
	"context"
	"net/http"
)

func HandleGrokPrompt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	url := "https://api.x.ai/v1/grok/completions"
	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API returned status: " + resp.Status)
}

	return ok("Grok responded to: " + prompt)
}