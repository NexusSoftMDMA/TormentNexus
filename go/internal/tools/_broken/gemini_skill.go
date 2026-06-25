package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	atlasBaseURL = "https://api.atlascloud.ai/v1"
)

var (
	atlasAPIKey = os.Getenv("ATLAS_API_KEY")
	atlasModel  = os.Getenv("ATLAS_MODEL")
)

func HandleAtlasListModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", atlasBaseURL+"/models", nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Authorization", "Bearer "+atlasAPIKey)
	resp, respErr := client.Do(req)
	if respErr != nil {
		return err(respErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error: %s", resp.Status))
}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	parseErr := json.NewDecoder(resp.Body).Decode(&result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	var models []string
	for _, model := range result.Data {
		models = append(models, model.ID)

	return ok(strings.Join(models, ", "))
}

}

func HandleAtlasSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	payload := map[string]interface{}{
		"model":     atlasModel,
		"messages":  []map[string]string{{"role": "user", "content": message}},
		"stream":    false,
		"max_tokens": 1024,
	}

	jsonPayload, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "POST", atlasBaseURL+"/chat/completions", strings.NewReader(string(jsonPayload)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Authorization", "Bearer "+atlasAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, respErr := client.Do(req)
	if respErr != nil {
		return err(respErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error: %s", resp.Status))
}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	parseErr := json.NewDecoder(resp.Body).Decode(&result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	if len(result.Choices) == 0 {
		return err("no response from model")
}

	return ok(result.Choices[0].Message.Content)
}

func HandleAtlasStreamMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	payload := map[string]interface{}{
		"model":    atlasModel,
		"messages": []map[string]string{{"role": "user", "content": message}},
		"stream":   true,
	}

	jsonPayload, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "POST", atlasBaseURL+"/chat/completions", strings.NewReader(string(jsonPayload)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Authorization", "Bearer "+atlasAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, respErr := client.Do(req)
	if respErr != nil {
		return err(respErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error: %s", resp.Status))
}

	reader := resp.Body
	buffer := make([]byte, 1024)
	var fullResponse strings.Builder

	for {
		n, readErr := reader.Read(buffer)
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return err(readErr.Error())
}

		data := string(buffer[:n])
		lines := strings.Split(data, "\n")

		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}

			parseErr := json.Unmarshal([]byte(line), &chunk)
			if parseErr != nil {
				return err(parseErr.Error())
}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				fullResponse.WriteString(chunk.Choices[0].Delta.Content)

		}
	}

	return ok(fullResponse.String())
}

}

func HandleGeminiGenerateImage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	newSession, _ :=getBool(args, "newSession", false)
	referenceImages, _ :=getString(args, "referenceImages", "")
	fullSize, _ :=getBool(args, "fullSize", true)
	timeout, _ :=getInt(args, "timeout", 120000)

	// TODO: Implement actual Gemini image generation logic
	// This is a placeholder implementation
	return ok(fmt.Sprintf("Generated image for prompt: %s", prompt))
}

func HandleGeminiNewChat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// TODO: Implement new chat session logic
	return ok("New chat session started")
}

func HandleGeminiSwitchModel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	if model == "" {
		return err("model is required")
}

	// TODO: Implement model switching logic
	return ok(fmt.Sprintf("Switched to model: %s", model))
}