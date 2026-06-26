package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleKnitGetMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	key, _ :=getString(args, "key")
	if project == "" || key == "" {
		return err("project and key are required")
}

	url := fmt.Sprintf("http://localhost:9876/memory/%s/%s", project, key)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return success(string(body))
}

func HandleKnitSetMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	if project == "" || key == "" || value == "" {
		return err("project, key, and value are required")
}

	payload := map[string]string{"value": value}
	data, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal payload: %v", e))
}

	url := fmt.Sprintf("http://localhost:9876/memory/%s/%s", project, key)
	req, e := http.NewRequestWithContext(ctx, "PUT", url, strings.NewReader(string(data)))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return success(string(body))
}// touch 1781132129
