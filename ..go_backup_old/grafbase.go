package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGrafbaseQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	apiKey := os.Getenv("GRAFBASE_API_KEY")
	if apiKey == "" {
		return err("GRAFBASE_API_KEY not set")
}

	apiURL := os.Getenv("GRAFBASE_API_URL")
	if apiURL == "" {
		apiURL = "https://api.grafbase.com/graphql"
	}
	payload := map[string]string{"query": query}
	body, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal request: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(respBody)))
}

	return ok(string(respBody))
}