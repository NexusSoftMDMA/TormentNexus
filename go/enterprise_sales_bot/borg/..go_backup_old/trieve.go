package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	apiKey, _ :=getString(args, "api_key")
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "https://api.trieve.ai"
	}
	url := fmt.Sprintf("%s/api/v1/search?query=%s", baseURL, query)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
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
		return err("API error: " + string(body))
}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return ok(fmt.Sprintf("Search results: %v", result))
}

func HandleGetDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	docID, _ :=getString(args, "document_id")
	apiKey, _ :=getString(args, "api_key")
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "https://api.trieve.ai"
	}
	url := fmt.Sprintf("%s/api/v1/document/%s", baseURL, docID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
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
		return err("API error: " + string(body))
}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return ok(fmt.Sprintf("Document: %v", result))
}