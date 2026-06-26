package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGenerateImage_mcp_server_aoai_dalle3(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	size, _ :=getString(args, "size")
	if size == "" {
		size = "1024x1024"
	}
	n, _ :=getInt(args, "n")
	if n == 0 {
		n = 1
	}

	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	if apiKey == "" || endpoint == "" {
		return err("AZURE_OPENAI_API_KEY and AZURE_OPENAI_ENDPOINT must be set")
}

	body := map[string]interface{}{
		"prompt": prompt,
		"size":   size,
		"n":      n,
	}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal request: %v", e))
}

	url := fmt.Sprintf("%s/openai/deployments/dall-e-3/images/generations?api-version=2024-02-01", endpoint)
	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	respBytes, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBytes)))
}

	var result struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if e := json.Unmarshal(respBytes, &result); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
}

	if len(result.Data) == 0 {
		return err("no images returned")
}

	imageURL := result.Data[0].URL
	return ok(fmt.Sprintf("Generated image: %s", imageURL))
}