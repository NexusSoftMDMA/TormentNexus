package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetFigmaFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileKey, _ :=getString(args, "file_key")
	if fileKey == "" {
		return err("file_key is required")
}

	base := os.Getenv("FIGMA_API_BASE_URL")
	if base == "" {
		base = "https://api.figma.com/v1"
	}
	token := os.Getenv("FIGMA_ACCESS_TOKEN")
	url := fmt.Sprintf("%s/files/%s", base, fileKey)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("X-Figma-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var data map[string]interface{}
	e = json.Unmarshal(body, &data)
	if e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	name, found := data["name"].(string)
	if !found {
		name = "untitled"
	}
	return success(fmt.Sprintf("File name: %s", name))
}

func HandleGenerateTwigsComponent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	nodeID, _ :=getString(args, "node_id")
	if nodeID == "" {
		return err("node_id is required")
}

	output := fmt.Sprintf("// Twigs component for node %s\nconst Component = () => {\n  return <View />;\n};", nodeID)
	return ok(output)
}