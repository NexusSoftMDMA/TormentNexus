package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleAiderExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	aiderArgs, _ :=getString(args, "args")
	body, _ := json.Marshal(map[string]string{"command": cmd, "args": aiderArgs})
	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/execute", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response failed: %v", e))
}

	return ok(string(data))
}

func HandleAiderStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/status", nil)
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response failed: %v", e))
}

	return success(string(data))
}