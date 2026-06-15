//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func HandleGenerateImage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return err("OPENAI_API_KEY not set")
}

	body := fmt.Sprintf(`{"model":"dall-e-3","prompt":"%s","n":1,"size":"1024x1024"}`, strings.ReplaceAll(prompt, `"`, `\"`))
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, "https://api..com/v1/images/generations", strings.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("api call failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Data []struct {
			URL string `json:"url"`,
		} `json:"data"`,
	}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("decode failed: " + e.Error())
	if len(result.Data) == 0 {
		return err("no image returned")
	return success("Image generated: " + result.Data[0].URL)
}
}
}