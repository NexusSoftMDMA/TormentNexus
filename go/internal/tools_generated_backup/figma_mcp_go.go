package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "file_key")
	if key == "" {
		return err("file_key is required")
}

	token := os.Getenv("FIGMA_ACCESS_TOKEN")
	if token == "" {
		return err("FIGMA_ACCESS_TOKEN not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.figma.com/v1/files/%s", key), nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return success(string(body))
}

func HandleGetImage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "file_key")
	id, _ :=getString(args, "node_id")
	if key == "" || id == "" {
		return err("file_key and node_id are required")
}

	token := os.Getenv("FIGMA_ACCESS_TOKEN")
	if token == "" {
		return err("FIGMA_ACCESS_TOKEN not set")
}

	url := fmt.Sprintf("https://api.figma.com/v1/images/%s?ids=%s", key, id)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return success(string(body))
}