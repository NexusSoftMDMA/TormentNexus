package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleMarkmapGenerate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	md, _ :=getString(args, "markdown")
	if md == "" {
		return err("markdown is required")
}

	body, _ := json.Marshal(map[string]string{"markdown": md})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.markmap.org/v1/render", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to call markmap API: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("API returned status " + resp.Status)
}

	var result struct{ SVG string `json:"svg"` }
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(result.SVG)
}