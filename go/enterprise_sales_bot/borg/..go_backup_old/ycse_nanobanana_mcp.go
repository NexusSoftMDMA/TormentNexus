package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleChat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	imageB64, _ :=getString(args, "image")
	model, _ :=getString(args, "model")
	if model == "" {
		model = "gemini-2.0-flash"
	}
	apiKey := os.Getenv("GEMINI_API_KEY")
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)

	contents := []map[string]interface{}{
		{"parts": []map[string]interface{}{}},
	}
	parts := []map[string]interface{}{}
	if imageB64 != "" {
		parts = append(parts, map[string]interface{}{
			"inline_data": map[string]interface{}{
				"mime_type": "image/jpeg",
				"data":      imageB64,
			},
		})

	parts = append(parts, map[string]interface{}{
		"text": prompt,
	})
	contents[0]["parts"] = parts

	reqBody := map[string]interface{}{
		"contents": contents,
	}
	body, e := json.Marshal(reqBody)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewReader(body))
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("api error: %s", result))
}

	candidates, found := result["candidates"].([]interface{})
	if !found || len(candidates) == 0 {
		return err("no candidates in response")
}

	candidate := candidates[0].(map[string]interface{})
	content, found := candidate["content"].(map[string]interface{})
	if !found {
		return err("no content in candidate")
}

	partsList, found := content["parts"].([]interface{})
	if !found || len(partsList) == 0 {
		return err("no parts in content")
}

	text, found := partsList[0].(map[string]interface{})["text"].(string)
	if !found {
		return err("no text in part")
}

	return ok(text)
}

}

func HandleImage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	model, _ :=getString(args, "model")
	if model == "" {
		model = "gemini-2.0-flash-exp-image-generation"
	}
	apiKey := os.Getenv("GEMINI_API_KEY")
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
	}
	body, e := json.Marshal(reqBody)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewReader(body))
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("api error: %s", result))
}

	candidates, found := result["candidates"].([]interface{})
	if !found || len(candidates) == 0 {
		return err("no candidates")
}

	candidate := candidates[0].(map[string]interface{})
	content, found := candidate["content"].(map[string]interface{})
	if !found {
		return err("no content")
}

	parts, found := content["parts"].([]interface{})
	if !found || len(parts) == 0 {
		return err("no parts")
}

	part := parts[0].(map[string]interface{})
	if inline, found := part["inlineData"].(map[string]interface{}); found {
		mimeType, _ := inline["mimeType"].(string)
		data, _ := inline["data"].(string)
		if mimeType == "" || data == "" {
			return err("invalid image data")
}

		return ok(fmt.Sprintf("data:%s;base64,%s", mimeType, data))
}

	// fallback to text
	text, found := part["text"].(string)
	if !found {
		return err("no image or text in response")
}

	return ok(text)
}