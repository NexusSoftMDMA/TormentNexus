package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// HandleLists retrieves Mailchimp audience lists.
func HandleLists(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	parts := strings.Split(apiKey, "-")
	if len(parts) != 2 {
		return err("invalid api_key format (expecting key-dc)")
}

	dc := parts[1]
	url := fmt.Sprintf("https://%s.api.mailchimp.com/3.0/lists", dc)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth("anystring", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %d %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json decode: " + e.Error())
}

	lists, found := result["lists"].([]interface{})
	if !found {
		return err("unexpected response format")
}

	return ok(fmt.Sprintf("Found %d lists", len(lists)))
}