package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleMemWrite(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	if key == "" || value == "" {
		return err("key and value are required")
}

	body, _ := json.Marshal(map[string]string{"key": key, "value": value})
	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/memento/mem_write", strings.NewReader(string(body)))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("memento request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("memento returned status %d", resp.StatusCode))
}

	return ok("memory written")
}