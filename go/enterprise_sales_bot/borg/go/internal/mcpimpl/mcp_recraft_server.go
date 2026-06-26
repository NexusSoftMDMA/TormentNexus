package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGenerateImage_mcp_recraft_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	body, _ := json.Marshal(map[string]string{"prompt": prompt})
	resp, e := http.DefaultClient.Post("https://api.recraft.ai/v1/images/generations", "application/json", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return ok(fmt.Sprintf("Image generated: %v", result["data"]))
}

func HandleEditImage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	image, _ :=getString(args, "image")
	prompt, _ :=getString(args, "prompt")
	if image == "" || prompt == "" {
		return err("image and prompt are required")
}

	body, _ := json.Marshal(map[string]string{"image": image, "prompt": prompt})
	resp, e := http.DefaultClient.Post("https://api.recraft.ai/v1/images/edits", "application/json", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return ok(fmt.Sprintf("Image edited: %v", result["data"]))
}