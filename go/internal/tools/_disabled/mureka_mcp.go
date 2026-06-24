package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type murekaResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
}

// HandleListModels returns available music generation models
func HandleListModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.mureka.ai/v1/models")
	if e != nil {
		return err("failed to fetch models: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response body")
}

	var mr murekaResponse
	if e := json.Unmarshal(body, &mr); e != nil {
		return err("failed to parse response")
}

	if !mr.Success {
		return err("API returned failure")
}

	return success(fmt.Sprintf("Models: %s", string(mr.Data)))
}

// HandleGenerateMusic generates music based on prompt and optional parameters
func HandleGenerateMusic(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	reqBody := map[string]interface{}{
		"prompt": prompt,
	}
	if v := getString(args, "style"); v != "" {
		reqBody["style"] = v
	}
	if v := getInt(args, "duration"); v > 0 {
		reqBody["duration"] = v
	}
	body, _ := json.Marshal(reqBody)
	resp, e := http.DefaultClient.Post("https://api.mureka.ai/v1/generate", "application/json", body)
	if e != nil {
		return err("failed to generate: " + e.Error())
}

	defer resp.Body.Close()
	rbody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var mr murekaResponse
	if e := json.Unmarshal(rbody, &mr); e != nil {
		return err("failed to parse response")
}

	if !mr.Success {
		return err("API returned failure")
}

	return success(fmt.Sprintf("Generated: %s", string(mr.Data)))
}