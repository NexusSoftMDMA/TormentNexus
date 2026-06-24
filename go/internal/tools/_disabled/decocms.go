package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleListConnections(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/connections", nil)
	if e != nil {
		return err("Failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Failed to fetch connections")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response")
}

	return ok(string(body))
}

func HandleCreateConnection(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	provider, _ :=getString(args, "provider")
	apiKey, _ :=getString(args, "apiKey")
	payload := map[string]string{"name": name, "provider": provider, "apiKey": apiKey}
	bodyBytes, e := json.Marshal(payload)
	if e != nil {
		return err("Failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/connections", bytes.NewReader(bodyBytes))
	if e != nil {
		return err("Failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Failed to create connection")
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response")
}

	return ok(string(respBody))
}