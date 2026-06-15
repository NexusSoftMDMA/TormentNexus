package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HandleListTopics fetches a list of topics from a Discourse instance.
func HandleListTopics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	apiKey, _ :=getString(args, "api_key")
	apiUser, _ :=getString(args, "api_username")
	limit, _ :=getInt(args, "limit")
	if base == "" {
		return err("base_url is required")
}

	url := fmt.Sprintf("%s/latest.json?limit=%d", base, limit)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	if apiKey != "" {
		req.Header.Set("Api-Key", apiKey)
		req.Header.Set("Api-Username", apiUser)

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
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}

}

// HandleGetTopic fetches a single topic by ID.
func HandleGetTopic(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	apiKey, _ :=getString(args, "api_key")
	apiUser, _ :=getString(args, "api_username")
	id, _ :=getInt(args, "topic_id")
	if base == "" || id == 0 {
		return err("base_url and topic_id are required")
}

	url := fmt.Sprintf("%s/t/%d.json", base, id)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	if apiKey != "" {
		req.Header.Set("Api-Key", apiKey)
		req.Header.Set("Api-Username", apiUser)

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
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(string(body))
}
}