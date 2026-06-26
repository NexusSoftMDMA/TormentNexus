package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type sentimentResponse struct {
	Label string `json:"label"`
	Score float64 `json:"score"`
}

func HandleAnalyzeSentiment_sentisift_sdks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text parameter is required")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.sentisift.com/v1/sentiment", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result sentimentResponse
	if e = json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
}

	return success(fmt.Sprintf("Sentiment: %s (score: %.2f)", result.Label, result.Score))
}

func HandleGetAvailableModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.sentisift.com/v1/models")
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var models []string
	if e = json.Unmarshal(body, &models); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
}

	return success(fmt.Sprintf("Available models: %v", models))
}