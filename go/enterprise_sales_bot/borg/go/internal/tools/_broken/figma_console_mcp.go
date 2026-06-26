package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetFile_figma_console_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileKey, _ :=getString(args, "fileKey")
	if fileKey == "" {
		return err("Missing fileKey")
}

	token, _ :=getString(args, "token")
	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.figma.com/v1/files/%s", fileKey), nil)
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("X-Figma-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return ok(string(body))
}

func HandleGetStyles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileKey, _ :=getString(args, "fileKey")
	if fileKey == "" {
		return err("Missing fileKey")
}

	token, _ :=getString(args, "token")
	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.figma.com/v1/files/%s/styles", fileKey), nil)
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("X-Figma-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return ok(string(body))
}