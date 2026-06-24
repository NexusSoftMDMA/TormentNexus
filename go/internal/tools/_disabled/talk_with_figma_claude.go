package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetFigmaFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileKey, _ :=getString(args, "file_key")
	if fileKey == "" {
		return err("file_key is required")
}

	token, _ :=getString(args, "token")
	url := fmt.Sprintf("https://api.figma.com/v1/files/%s", fileKey)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	req.Header.Set("X-Figma-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode response: %v", e))
}

	return success(fmt.Sprintf("Figma file: %v", result["name"]))
}