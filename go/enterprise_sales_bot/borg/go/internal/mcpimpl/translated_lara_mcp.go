package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleTranslate_translated_lara_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	source, _ :=getString(args, "source_lang")
	target, _ :=getString(args, "target_lang")
	apiKey, _ :=getString(args, "api_key")
	if text == "" || target == "" {
		return err("Missing required parameters: text and target_lang")
}

	if apiKey == "" {
		return err("Missing required parameter: api_key")
}

	payload := map[string]interface{}{
		"text":        text,
		"source_lang": source,
		"target_lang": target,
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("Failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.lara.translated.com/translate", strings.NewReader(string(body)))
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("Failed to parse response: " + e.Error())
}

	translated, found := result["translated_text"].(string)
	if !found {
		return err("Unexpected response format")
}

	return ok(translated)
}

func HandleLanguages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("Missing required parameter: api_key")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.lara.translated.com/languages", nil)
	if e != nil {
		return err("Failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("Request failed: " + e.Error())
}

	defer resp.Body.Close()
	var languages []string
	if e := json.NewDecoder(resp.Body).Decode(&languages); e != nil {
		return err("Failed to parse response: " + e.Error())
}

	return ok(strings.Join(languages, ", "))
}