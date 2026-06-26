package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleReadMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("key is required")
}

	url := fmt.Sprintf("http://localhost:9090/memory/%s", key)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("daemon request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("daemon error: %s", string(body)))
}

	return ok(string(body))
}

func HandleWriteMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	if key == "" || value == "" {
		return err("key and value are required")
}

	payload := map[string]string{"key": key, "value": value}
	data, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("marshal failed: %v", e))
}

	url := "http://localhost:9090/memory"
	resp, e := http.DefaultClient.Post(url, "application/json", strings.NewReader(string(data)))
	if e != nil {
		return err(fmt.Sprintf("daemon request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("daemon error: %s", string(body)))
}

	return ok("written")
}