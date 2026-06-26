package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandlePatchContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	content, _ :=getString(args, "content")

	if strings.Contains(path, "..") {
		return err("path contains insecure sequence '..'")
}

	if path == "" {
		return err("path is required")
}

	body, e := json.Marshal(map[string]string{"path": path, "content": content})
	if e != nil {
		return err("failed to marshal request body")
}

	resp, e := http.DefaultClient.Post("http://localhost:27124/patchContent", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to send request: " + e.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return err("upstream returned status " + resp.Status)
}

	return ok("content patched successfully")
}