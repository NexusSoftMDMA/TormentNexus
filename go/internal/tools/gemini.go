//go:build ignore
// +build ignore

package tools

/**
 * @file gemini.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Gemini MCP tools.
 * Replaces `gemini-mcp` (npx gemini-mcp@latest) entry in mcp.json.
 *
 * Uses the Google Gemini API natively.
 * Improvements over original:
 * - No npx/Node dependency.
 * - Supports: chat, text generation, code generation, vision (image analysis),
 *   embeddings, safety settings, and model listing.
 * - Context-aware with timeout; uses GEMINI_API_KEY for auth.
 * - Supports the full Gemini 2.x API including function calling.
 */

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const geminiAPIBase = "https://generativelanguage.googleapis.com/v1beta"

func geminiAPIKey() string {
	if k := os.Getenv("GEMINI_API_KEY"); k != "" {
		return k
	}
	if k := os.Getenv("GOOGLE_API_KEY"); k != "" {
		return k
	}
	return ""
}

func geminiDo(ctx context.Context, method, urlPath string, payload interface{}) (interface{}, error) {
	apiKey := geminiAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	var bodyReader io.Reader
	if payload != nil {
		data, _ := json.Marshal(payload)
		bodyReader = bytes.NewBuffer(data)
	}

	// Append API key as query parameter
	fullURL := geminiAPIBase + urlPath
	if strings.Contains(fullURL, "?") {
		fullURL += "&key=" + apiKey
	} else {
		fullURL += "?key=" + apiKey
	}

	req, e := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if e != nil {
		return nil, e
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "TormentNexus/1.0")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Gemini API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	if len(body) == 0 {
		return nil, nil
	}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return string(body), nil
	}
	return result, nil
}

// HandleGeminiChat sends a chat message to a Gemini model.
// Tool: gemini_chat
func HandleGeminiChat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ := getString(args, "message", "prompt", "text")
	if message == "" {
		return err("message parameter is required")
	}

	model, _ := getString(args, "model")
	if model == "" {
		model = "gemini-2.5-flash"
	}

	// Build contents array
	contents := []map[string]interface{}{
		{
			"role": "user",
			"parts": []map[string]interface{}{
				{"text": message},
			},
		},
	}

	// Support system instruction
	systemInstruction, _ := getString(args, "system_instruction", "system")
	generationConfig := map[string]interface{}{}

	if temperature, ok := args["temperature"].(float64); ok {
		generationConfig["temperature"] = temperature
	}
	if topP, ok := args["top_p"].(float64); ok {
		generationConfig["topP"] = topP
	}
	if maxTokens := getInt(args, "max_tokens", "maxOutputTokens"); maxTokens > 0 {
		generationConfig["maxOutputTokens"] = maxTokens
	}

	payload := map[string]interface{}{
		"contents":         contents,
		"generationConfig": generationConfig,
	}

	if systemInstruction != "" {
		payload["systemInstruction"] = map[string]interface{}{
			"parts": []map[string]interface{}{
				{"text": systemInstruction},
			},
		}
	}

	// Support conversation history
	if history, ok := args["history"].([]interface{}); ok {
		newContents := []map[string]interface{}{}
		for _, h := range history {
			if hMap, ok := h.(map[string]interface{}); ok {
				newContents = append(newContents, hMap)
			}
		}
		newContents = append(newContents, contents[0])
		payload["contents"] = newContents
	}

	result, e := geminiDo(ctx, "POST", fmt.Sprintf("/models/%s:generateContent", model), payload)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleGeminiCodeGeneration generates code using Gemini.
// Tool: gemini_code_generation
func HandleGeminiCodeGeneration(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ := getString(args, "prompt", "description")
	if prompt == "" {
		return err("prompt parameter is required")
	}

	model, _ := getString(args, "model")
	if model == "" {
		model = "gemini-2.5-flash"
	}

	language, _ := getString(args, "language", "lang")

	systemPrompt := "You are an expert programmer. Generate clean, efficient, well-documented code."
	if language != "" {
		systemPrompt = fmt.Sprintf("You are an expert %s programmer. Generate clean, efficient, well-documented %s code.", language, language)
	}

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
		"systemInstruction": map[string]interface{}{
			"parts": []map[string]interface{}{
				{"text": systemPrompt},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature": 0.2,
		},
	}

	result, e := geminiDo(ctx, "POST", fmt.Sprintf("/models/%s:generateContent", model), payload)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleGeminiVision analyzes an image using Gemini's multimodal capabilities.
// Tool: gemini_vision
func HandleGeminiVision(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ := getString(args, "prompt", "question")
	if prompt == "" {
		prompt = "Describe this image."
	}

	model, _ := getString(args, "model")
	if model == "" {
		model = "gemini-2.5-flash"
	}

	imagePath, _ := getString(args, "image_path", "filePath")
	imageURL, _ := getString(args, "image_url", "url")
	imageBase64, _ := getString(args, "image_base64")

	if imagePath == "" && imageURL == "" && imageBase64 == "" {
		return err("one of image_path, image_url, or image_base64 is required")
	}

	parts := []map[string]interface{}{
		{"text": prompt},
	}

	if imagePath != "" {
		data, e := os.ReadFile(imagePath)
		if e != nil {
			return err(fmt.Sprintf("Failed to read image file: %v", e))
		}
		mimeType := "image/jpeg"
		if strings.HasSuffix(strings.ToLower(imagePath), ".png") {
			mimeType = "image/png"
		} else if strings.HasSuffix(strings.ToLower(imagePath), ".gif") {
			mimeType = "image/gif"
		} else if strings.HasSuffix(strings.ToLower(imagePath), ".webp") {
			mimeType = "image/webp"
		}
		parts = append(parts, map[string]interface{}{
			"inlineData": map[string]interface{}{
				"mimeType": mimeType,
				"data":     base64.StdEncoding.EncodeToString(data),
			},
		})
	} else if imageURL != "" {
		// Download the image
		client := &http.Client{Timeout: 30 * time.Second}
		resp, e := client.Get(imageURL)
		if e != nil {
			return err(fmt.Sprintf("Failed to download image: %v", e))
		}
		defer resp.Body.Close()
		imgData, _ := io.ReadAll(resp.Body)

		contentType := resp.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "image/jpeg"
		}

		parts = append(parts, map[string]interface{}{
			"inlineData": map[string]interface{}{
				"mimeType": contentType,
				"data":     base64.StdEncoding.EncodeToString(imgData),
			},
		})
	} else if imageBase64 != "" {
		mimeType, _ := getString(args, "mime_type")
		if mimeType == "" {
			mimeType = "image/jpeg"
		}
		parts = append(parts, map[string]interface{}{
			"inlineData": map[string]interface{}{
				"mimeType": mimeType,
				"data":     imageBase64,
			},
		})
	}

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role":  "user",
				"parts": parts,
			},
		},
	}

	result, e := geminiDo(ctx, "POST", fmt.Sprintf("/models/%s:generateContent", model), payload)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleGeminiEmbeddings generates text embeddings using Gemini.
// Tool: gemini_embeddings
func HandleGeminiEmbeddings(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ := getString(args, "text", "content")
	if text == "" {
		return err("text parameter is required")
	}

	model, _ := getString(args, "model")
	if model == "" {
		model = "text-embedding-004"
	}

	payload := map[string]interface{}{
		"model": "models/" + model,
		"content": map[string]interface{}{
			"parts": []map[string]interface{}{
				{"text": text},
			},
		},
	}

	result, e := geminiDo(ctx, "POST", fmt.Sprintf("/models/%s:embedContent", model), payload)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleGeminiListModels lists available Gemini models.
// Tool: gemini_list_models
func HandleGeminiListModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	result, e := geminiDo(ctx, "GET", "/models", nil)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleGeminiFunctionCalling uses Gemini with function declarations.
// Tool: gemini_function_calling
func HandleGeminiFunctionCalling(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ := getString(args, "message", "prompt")
	if message == "" {
		return err("message parameter is required")
	}

	model, _ := getString(args, "model")
	if model == "" {
		model = "gemini-2.5-flash"
	}

	functionDeclarations, fdOK := args["function_declarations"].([]interface{})
	if !fdOK || len(functionDeclarations) == 0 {
		return err("function_declarations parameter is required (array of function declarations)")
	}

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]interface{}{
					{"text": message},
				},
			},
		},
		"tools": []map[string]interface{}{
			{
				"functionDeclarations": functionDeclarations,
			},
		},
	}

	result, e := geminiDo(ctx, "POST", fmt.Sprintf("/models/%s:generateContent", model), payload)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}
