package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleReadSecret(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	url := "https://vault.example.com/v1/secret/data/" + path
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("X-Vault-Token", getString(args, "token"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read body: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("vault returned status %d: %s", resp.StatusCode, string(body)))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
}

	secret, found := data["data"].(map[string]interface{})["data"].(map[string]interface{})["data"]
	if !found {
		return err("no secret data found")
}

	return success(fmt.Sprintf("secret: %v", secret))
}