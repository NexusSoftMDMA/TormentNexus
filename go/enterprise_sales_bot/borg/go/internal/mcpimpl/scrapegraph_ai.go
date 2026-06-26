package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleScrapegraph(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	url, _ :=getString(args, "url")
	if apiKey == "" || url == "" {
		return err("api_key and url are required")
}

	body, e := json.Marshal(map[string]string{"url": url})
	if e != nil {
		return err("failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.scrapegraphai.com/v1/scrape", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("failed to decode response")
}

	return ok(fmt.Sprintf("Scraped content: %v", result["data"]))
}