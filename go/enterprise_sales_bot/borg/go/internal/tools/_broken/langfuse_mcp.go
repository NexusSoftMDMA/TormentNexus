package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func HandleCreateTrace(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	userID, _ :=getString(args, "userId")
	tags, _ :=getString(args, "tags")
	baseURL := os.Getenv("LANGFUSE_BASE_URL")
	if baseURL == "" {
		return err("LANGFUSE_BASE_URL environment variable not set")
}

	body := map[string]interface{}{"name": name}
	if userID != "" {
		body["userId"] = userID
	}
	if tags != "" {
		body["tags"] = strings.Split(tags, ",")

	jsonBytes, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", baseURL+"/api/public/traces", strings.NewReader(string(jsonBytes)))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	publicKey := os.Getenv("LANGFUSE_PUBLIC_KEY")
	secretKey := os.Getenv("LANGFUSE_SECRET_KEY")
	if publicKey != "" && secretKey != "" {
		req.SetBasicAuth(publicKey, secretKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	return ok("Trace created successfully")
}
}
}