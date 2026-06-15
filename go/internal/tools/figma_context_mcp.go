//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetFileInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileKey, _ :=getString(args, "fileKey")
	if fileKey == "" {
		return err("fileKey is required")
}

	token := os.Getenv("FIGMA_ACCESS_TOKEN")
	if token == "" {
		return err("FIGMA_ACCESS_TOKEN not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.figma.com/v1/files/%s", fileKey), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	name, found := result["name"].(string)
	if !found {
		name = "unknown"
	}
	return ok(fmt.Sprintf("File name: %s", name))
}
