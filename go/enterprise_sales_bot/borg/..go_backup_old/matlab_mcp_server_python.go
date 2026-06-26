package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleExecuteMatlab(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return ok("No command provided")
}

	payload := map[string]string{"command": cmd}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("Failed to marshal: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:5000/execute", bytes.NewReader(body))
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBytes, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	return ok(string(respBytes))
}