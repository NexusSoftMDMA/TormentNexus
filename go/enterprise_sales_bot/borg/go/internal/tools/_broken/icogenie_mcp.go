package mcpimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// HandleGenerateIcon generates an icon based on a prompt.
func HandleGenerateIcon(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	url := "https://api.icogenie.com/generate"
	body := map[string]string{"prompt": prompt}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, io.NopCloser(bytes.NewReader(jsonBody)))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("API error: " + string(respBody))
}

	return ok("Icon generated: " + string(respBody))
}