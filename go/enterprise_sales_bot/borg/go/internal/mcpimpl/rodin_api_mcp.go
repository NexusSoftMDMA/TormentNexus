package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleGenerateImage_rodin_api_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	model, _ :=getString(args, "model")
	size, _ :=getString(args, "size")
	if prompt == "" {
		return err("prompt is required")
}

	body := fmt.Sprintf(`{"prompt":"%s","model":"%s","size":"%s"}`, prompt, model, size)
	resp, e := http.DefaultClient.Post("https://api.rodin.ai/v1/images/generations", "application/json", strings.NewReader(body))
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(b)))
}

	var result map[string]interface{}
	json.Unmarshal(b, &result)
	return ok(fmt.Sprintf("Generated image: %v", result["data"]))
}

func HandleListModels_rodin_api_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.rodin.ai/v1/models")
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(b)))
}

	var result map[string]interface{}
	json.Unmarshal(b, &result)
	return ok(fmt.Sprintf("Models: %v", result["data"]))
}